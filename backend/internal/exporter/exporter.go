package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"getreleased/internal/database"
)

const recentReleasesLimit = 50

func Export(ctx context.Context, db *database.DB, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	repos, err := db.ListRepositories(ctx)
	if err != nil {
		return err
	}

	releasesDir := filepath.Join(dir, "releases")
	if err := os.MkdirAll(releasesDir, 0o755); err != nil {
		return err
	}

	for i := range repos {
		repoReleases, err := db.GetReleasesByRepository(ctx, repos[i].ID)
		if err != nil {
			return err
		}
		repos[i].ReleaseCount = len(repoReleases)
		if len(repoReleases) > 0 {
			repos[i].LatestIsPrerelease = repoReleases[0].IsPrerelease
		}
		if err := writeJSON(filepath.Join(releasesDir, strconv.FormatInt(repos[i].ID, 10)+".json"), repoReleases); err != nil {
			return err
		}
	}

	if err := writeJSON(filepath.Join(dir, "repositories.json"), repos); err != nil {
		return err
	}

	recent, err := db.ListRecentReleases(ctx, recentReleasesLimit)
	if err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "releases-recent.json"), recent)
}

func writeJSON(path string, v any) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := errors.Join(enc.Encode(v), tmp.Close()); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}
