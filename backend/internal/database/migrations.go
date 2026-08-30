package database

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schema string

//go:embed seed_tags.sql
var seedTags string

func (d *DB) Migrate(ctx context.Context) error {
	if _, err := d.conn.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if err := d.migrateReleasesColumns(ctx); err != nil {
		return fmt.Errorf("migrate releases columns: %w", err)
	}

	if err := d.migrateRepositoriesColumns(ctx); err != nil {
		return fmt.Errorf("migrate repositories columns: %w", err)
	}

	return nil
}

// migrateReleasesColumns 为已存在的 releases 表补齐新增列。
// schema.sql 用 CREATE TABLE IF NOT EXISTS，对已建表加列不生效，故此处幂等 ALTER。
func (d *DB) migrateReleasesColumns(ctx context.Context) error {
	existing, err := d.tableColumns(ctx, "releases")
	if err != nil {
		return err
	}
	for _, col := range []string{"tarball_url", "zipball_url"} {
		if existing[col] {
			continue
		}
		if _, err := d.conn.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE releases ADD COLUMN %s TEXT NOT NULL DEFAULT ''`, col)); err != nil {
			return fmt.Errorf("add column %s: %w", col, err)
		}
	}
	if _, err := d.conn.ExecContext(ctx, `UPDATE releases SET tarball_url = '', zipball_url = '' WHERE tarball_url IS NULL OR zipball_url IS NULL`); err != nil {
		return fmt.Errorf("normalize releases nulls: %w", err)
	}
	return nil
}

// migrateRepositoriesColumns 为已存在的 repositories 表补齐 Phase 2+3 新增列。
// schema.sql 用 CREATE TABLE IF NOT EXISTS，对已建表加列不生效，故此处幂等 ALTER。
func (d *DB) migrateRepositoriesColumns(ctx context.Context) error {
	existing, err := d.tableColumns(ctx, "repositories")
	if err != nil {
		return err
	}
	additions := []struct {
		col string
		sql string
	}{
		{"pushed_at", `ALTER TABLE repositories ADD COLUMN pushed_at DATETIME`},
		{"etag", `ALTER TABLE repositories ADD COLUMN etag TEXT`},
		{"last_modified", `ALTER TABLE repositories ADD COLUMN last_modified TEXT`},
	}
	for _, a := range additions {
		if existing[a.col] {
			continue
		}
		if _, err := d.conn.ExecContext(ctx, a.sql); err != nil {
			return fmt.Errorf("add column %s: %w", a.col, err)
		}
	}
	return nil
}

func (d *DB) tableColumns(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := d.conn.QueryxContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return nil, fmt.Errorf("table_info %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()
	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan table_info: %w", err)
		}
		cols[name] = true
	}
	return cols, rows.Err()
}

func (d *DB) SeedTags(ctx context.Context) error {
	if _, err := d.conn.ExecContext(ctx, seedTags); err != nil {
		return fmt.Errorf("seed tags: %w", err)
	}
	return nil
}
