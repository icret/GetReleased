package exporter

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"getreleased/internal/database"
	"getreleased/internal/release"
)

func TestExportWritesAllJSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.OpenWithMaxConns(dbPath, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	repo := &release.Repository{Owner: "o", Name: "n", FullName: "o/n"}
	repoID, err := db.SaveRepository(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}

	exportDir := filepath.Join(dir, "out")
	if err := Export(context.Background(), db, exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}

	for _, name := range []string{"repositories.json", "releases-recent.json"} {
		info, err := os.Stat(filepath.Join(exportDir, name))
		if err != nil {
			t.Errorf("expected %s: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", name)
		}
	}

	shard := filepath.Join(exportDir, "releases", strconv.FormatInt(repoID, 10)+".json")
	info, err := os.Stat(shard)
	if err != nil {
		t.Fatalf("expected release shard %s: %v", shard, err)
	}
	if info.Size() == 0 {
		t.Errorf("%s is empty", shard)
	}
}
