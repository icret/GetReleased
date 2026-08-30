package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// TestMigrateReleasesNullsNormalized 验证旧库升级场景：
// releases 表的 tarball_url/zipball_url 列已存在且含 NULL 时，
// Migrate 应将其规范化为空字符串，避免 string 字段 scan NULL 报错。
func TestMigrateReleasesNullsNormalized(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	// 模拟旧库：releases 表已存在且 tarball_url/zipball_url 为可空列。
	if _, err := db.conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS releases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repository_id INTEGER NOT NULL,
		tag_name TEXT NOT NULL,
		name TEXT,
		body TEXT,
		html_url TEXT,
		tarball_url TEXT,
		zipball_url TEXT,
		published_at DATETIME,
		is_prerelease INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(repository_id, tag_name)
	)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.conn.ExecContext(ctx, `INSERT INTO releases (repository_id, tag_name, tarball_url, zipball_url) VALUES (1, 'v1.0.0', NULL, NULL)`); err != nil {
		t.Fatalf("seed null row: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var tar, zip string
	if err := db.conn.GetContext(ctx, &tar, `SELECT tarball_url FROM releases WHERE tag_name = 'v1.0.0'`); err != nil {
		t.Fatalf("select tarball: %v", err)
	}
	if err := db.conn.GetContext(ctx, &zip, `SELECT zipball_url FROM releases WHERE tag_name = 'v1.0.0'`); err != nil {
		t.Fatalf("select zipball: %v", err)
	}
	if tar != "" || zip != "" {
		t.Errorf("nulls not normalized: tarball=%q zipball=%q, want both empty", tar, zip)
	}
}

func TestMigrateRepositoriesNewColumns(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	if _, err := db.conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT NOT NULL,
		name TEXT NOT NULL,
		full_name TEXT NOT NULL UNIQUE,
		description TEXT,
		logo_path TEXT,
		stars INTEGER NOT NULL DEFAULT 0,
		language TEXT,
		is_archived INTEGER NOT NULL DEFAULT 0,
		is_private INTEGER NOT NULL DEFAULT 0,
		latest_version TEXT,
		latest_release_url TEXT,
		latest_release_date DATETIME,
		last_checked_at DATETIME,
		remark TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if _, err := db.conn.ExecContext(ctx, `INSERT INTO repositories (owner, name, full_name) VALUES ('o', 'n', 'o/n')`); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cols, err := db.tableColumns(ctx, "repositories")
	if err != nil {
		t.Fatalf("table_info: %v", err)
	}
	for _, col := range []string{"pushed_at", "etag", "last_modified"} {
		if !cols[col] {
			t.Errorf("column %s not added by migration", col)
		}
	}

	var etag, lastMod sql.NullString
	var pushedAt sql.NullTime
	if err := db.conn.QueryRowxContext(ctx, `SELECT etag, last_modified, pushed_at FROM repositories WHERE full_name = 'o/n'`).Scan(&etag, &lastMod, &pushedAt); err != nil {
		t.Fatalf("select new columns: %v", err)
	}
	if etag.Valid || lastMod.Valid {
		t.Errorf("new columns should be null, got etag=%q last_modified=%q", etag.String, lastMod.String)
	}
}
