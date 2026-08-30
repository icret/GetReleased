package main

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"getreleased/internal/api"
	"getreleased/internal/auth"
	"getreleased/internal/avatar"
	"getreleased/internal/database"
	"getreleased/internal/github"
	"getreleased/internal/logging"
	"getreleased/internal/tracker"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.ErrorContext(context.Background(), "load .env", "err", err)
		os.Exit(1)
	}

	isDev := os.Getenv("SERVER_DEV") == "true"
	slog.SetDefault(slog.New(logging.NewHandler(isDev)))

	dbPath := getenv("DB_PATH", "./backend/data/tracker.db")
	exportDir := getenv("EXPORT_DIR", "./frontend/public/data")
	avatarDir := getenv("AVATAR_DIR", "./frontend/public/assets/images/repos")
	addr := getenv("SERVER_ADDR", ":8080")

	jwtSecret := os.Getenv("JWT_SECRET")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		slog.ErrorContext(context.Background(), "mkdir db path", "err", err)
		os.Exit(1)
	}

	db, err := database.OpenWithMaxConns(dbPath, 5, 5)
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

	authSvc, err := auth.NewService(db, jwtSecret, 12*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "auth service", "err", err)
		os.Exit(1)
	}
	adminUser := getenv("ADMIN_USERNAME", "admin")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if err := authSvc.EnsureAdminSeed(ctx, adminUser, adminPass); err != nil {
		slog.ErrorContext(ctx, "seed admin", "err", err)
		os.Exit(1)
	}

	ghClient := github.NewClient(loadTokens())
	trk := tracker.New(db, ghClient, avatar.NewDownloader(avatarDir))
	apiHandler := api.New(db, trk, ghClient, authSvc, exportDir)

	server := &http.Server{
		Addr:              addr,
		Handler:           apiHandler.Router(isDev),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.InfoContext(ctx, "server listening", "addr", addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "shutdown", "err", err)
	}
}

func getenv(key, def string) string {
	return cmp.Or(os.Getenv(key), def)
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
