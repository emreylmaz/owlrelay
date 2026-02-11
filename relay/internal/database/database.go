// Package database handles SQLite database operations
package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// Schema for the database
const schema = `
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    rate_limit INTEGER NOT NULL DEFAULT 100,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TEXT,
    revoked_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_tokens_hash ON tokens(hash);
CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked_at);
`

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply schema
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}

	log.Debug().Str("path", dbPath).Msg("Database initialized")

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
