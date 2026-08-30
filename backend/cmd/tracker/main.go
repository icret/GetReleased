package main

import (
	"cmp"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"getreleased/internal/avatar"
	"getreleased/internal/database"
	"getreleased/internal/exporter"
	"getreleased/internal/github"
	"getreleased/internal/logging"
	"getreleased/internal/scheduler"
	"getreleased/internal/tracker"

	"github.com/joho/godotenv"
)

type trackedRepo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		slog.ErrorContext(context.Background(), "load .env", "err", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(logging.NewHandler(false)))

	dbPath := getenv("DB_PATH", "./backend/data/tracker.db")
	exportDir := getenv("EXPORT_DIR", "./frontend/public/data")
	avatarDir := getenv("AVATAR_DIR", "./frontend/public/assets/images/repos")
	reposFile := getenv("REPOS_FILE", "./backend/config/repositories.json")
	interval := getenvDuration("TRACK_INTERVAL", 30*time.Minute)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		slog.ErrorContext(context.Background(), "mkdir db path", "err", err)
		os.Exit(1)
	}

	db, err := database.Open(dbPath)
	if err != nil {
		slog.ErrorContext(context.Background(), "open db", "err", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		slog.ErrorContext(ctx, "migrate", "err", err)
		os.Exit(1)
	}
	if err := db.SeedTags(ctx); err != nil {
		slog.ErrorContext(ctx, "seed tags", "err", err)
		os.Exit(1)
	}

	if err := seedIfEmpty(ctx, db, reposFile); err != nil {
		slog.ErrorContext(ctx, "seed repositories", "err", err)
		os.Exit(1)
	}

	ghClient := github.NewClient(loadTokens())
	trk := tracker.New(db, ghClient, avatar.NewDownloader(avatarDir))

	task := func(ctx context.Context) {
		if remaining, err := ghClient.RateLimitRemaining(ctx); err != nil {
			slog.WarnContext(ctx, "rate limit check", "err", err)
		} else if remaining < 500 {
			slog.WarnContext(ctx, "rate limit low, skip track", "remaining", remaining)
			return
		}

		dirty, err := trk.Track(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "track", "err", err)
		}
		if !dirty {
			slog.InfoContext(ctx, "track done, no change, skip export")
			return
		}
		if err := exporter.Export(ctx, db, exportDir); err != nil {
			slog.ErrorContext(ctx, "export", "err", err)
		}
	}

	task(ctx)
	scheduler.New(interval, task).Start(ctx)
}

func loadTokens() []string {
	if v := os.Getenv("GITHUB_TOKENS"); v != "" {
		parts := strings.Split(v, ",")
		tokens := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				tokens = append(tokens, p)
			}
		}
		if len(tokens) > 0 {
			return tokens
		}
	}
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		return []string{v}
	}
	return nil
}

func seedIfEmpty(ctx context.Context, db *database.DB, reposFile string) error {
	count, err := db.CountRepositories(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seeds, err := loadSeedRepositories(reposFile)
	if err != nil {
		return err
	}
	if err := db.SeedRepositories(ctx, seeds); err != nil {
		return err
	}
	slog.InfoContext(ctx, "seeded repositories", "count", len(seeds), "file", reposFile)
	return nil
}

func loadSeedRepositories(path string) ([]database.RepoSeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var repos []trackedRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	seeds := make([]database.RepoSeed, len(repos))
	for i, r := range repos {
		seeds[i] = database.RepoSeed{Owner: r.Owner, Name: r.Name}
	}
	return seeds, nil
}

func getenv(key, def string) string {
	return cmp.Or(os.Getenv(key), def)
}

func getenvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if d, err := time.ParseDuration(v + "m"); err == nil {
		return d
	}
	return def
}
