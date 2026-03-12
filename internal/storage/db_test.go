/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("DB file should exist: %v", err)
	}
}

func TestOpenMigratesSchema(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// pending_urls table should exist after migration
	var name string
	err = db.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='pending_urls'`,
	).Scan(&name)
	if err != nil {
		t.Fatalf("pending_urls table should exist: %v", err)
	}
}

func TestOpenIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer db2.Close()
}

func TestCloseIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("second Close should not error: %v", err)
	}
}

func TestSchemaHasTimestamps(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	rows, err := db.db.Query(`PRAGMA table_info(pending_urls)`)
	if err != nil {
		t.Fatalf("PRAGMA: %v", err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt *string
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan: %v", err)
		}
		cols[name] = true
	}

	for _, want := range []string{"created_at", "updated_at"} {
		if !cols[want] {
			t.Errorf("pending_urls should have %q column", want)
		}
	}
}
