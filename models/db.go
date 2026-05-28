package models

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func Open(migrationsFS fs.FS) (*sqlx.DB, error) {
	path := DBPath()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sqlx.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(db.DB, "."); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func DBPath() string {
	if p := os.Getenv("DATABASE_PATH"); p != "" {
		return p
	}
	return "./data/party.db"
}
