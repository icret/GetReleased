package database

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	conn *sqlx.DB
}

func Open(path string) (*DB, error) {
	return open(path, 1, 1)
}

func OpenWithMaxConns(path string, maxOpen, maxIdle int) (*DB, error) {
	return open(path, maxOpen, maxIdle)
}

func open(path string, maxOpen, maxIdle int) (*DB, error) {
	conn, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
		`PRAGMA cache_size=-10000`,
		`PRAGMA busy_timeout=5000`,
		`PRAGMA foreign_keys=ON`,
	}
	for _, p := range pragmas {
		if _, err := conn.ExecContext(context.Background(), p); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("exec %q: %w", p, err)
		}
	}
	conn.SetMaxOpenConns(maxOpen)
	conn.SetMaxIdleConns(maxIdle)
	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) Conn() *sqlx.DB {
	return d.conn
}

func (d *DB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return d.conn.BeginTxx(ctx, nil)
}

func WithTransaction(ctx context.Context, db *DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
