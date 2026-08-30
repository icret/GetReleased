package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"getreleased/internal/database"
)

func Export(ctx context.Context, db *database.DB, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	repos, err := db.ListRepositories(ctx)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "repositories.json"), repos); err != nil {
		return err
	}

	releases, err := db.ListReleases(ctx)
	if err != nil {
		return err
	}
	return writeJSON(filepath.Join(dir, "releases.json"), releases)
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
