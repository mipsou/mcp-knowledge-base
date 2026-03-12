/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package storage

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite connection with schema migration.
type DB struct {
	db     *sql.DB
	closed bool
	mu     sync.Mutex
}

// Open opens (or creates) a SQLite database and runs migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// WAL mode for concurrent reads.
	if _, err := conn.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		conn.Close()
		return nil, err
	}

	d := &DB{db: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

// Close closes the database connection. Safe to call multiple times.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	return d.db.Close()
}

// migrate runs schema migrations.
func (d *DB) migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS pending_urls (
			id         TEXT PRIMARY KEY,
			url        TEXT NOT NULL,
			collection TEXT NOT NULL,
			status     TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}
