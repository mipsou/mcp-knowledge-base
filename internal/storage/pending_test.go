/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package storage

import (
	"path/filepath"
	"testing"

	"github.com/mipsou/mcp-biblium/internal/pending"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestPendingAdd(t *testing.T) {
	db := openTestDB(t)

	entry, err := db.PendingAdd("infra", "https://example.com/doc")
	if err != nil {
		t.Fatalf("PendingAdd: %v", err)
	}
	if entry.ID == "" {
		t.Error("expected non-empty ID")
	}
	if entry.URL != "https://example.com/doc" {
		t.Errorf("URL = %q, want %q", entry.URL, "https://example.com/doc")
	}
	if entry.Collection != "infra" {
		t.Errorf("Collection = %q, want %q", entry.Collection, "infra")
	}
	if entry.Status != pending.StatusPending {
		t.Errorf("Status = %q, want %q", entry.Status, pending.StatusPending)
	}
}

func TestPendingList(t *testing.T) {
	db := openTestDB(t)

	_, _ = db.PendingAdd("infra", "https://example.com/a")
	_, _ = db.PendingAdd("infra", "https://example.com/b")

	entries, err := db.PendingList()
	if err != nil {
		t.Fatalf("PendingList: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestPendingApprove(t *testing.T) {
	db := openTestDB(t)

	entry, _ := db.PendingAdd("infra", "https://example.com/doc")
	approved, err := db.PendingApprove(entry.ID)
	if err != nil {
		t.Fatalf("PendingApprove: %v", err)
	}
	if approved.Status != pending.StatusApproved {
		t.Errorf("Status = %q, want %q", approved.Status, pending.StatusApproved)
	}
}

func TestPendingReject(t *testing.T) {
	db := openTestDB(t)

	entry, _ := db.PendingAdd("infra", "https://example.com/doc")
	rejected, err := db.PendingReject(entry.ID)
	if err != nil {
		t.Fatalf("PendingReject: %v", err)
	}
	if rejected.Status != pending.StatusRejected {
		t.Errorf("Status = %q, want %q", rejected.Status, pending.StatusRejected)
	}
}

func TestPendingApproveUnknown(t *testing.T) {
	db := openTestDB(t)

	_, err := db.PendingApprove("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
}

func TestPendingRejectUnknown(t *testing.T) {
	db := openTestDB(t)

	_, err := db.PendingReject("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
}

func TestPendingSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db1, _ := Open(path)
	_, _ = db1.PendingAdd("infra", "https://example.com/persist")
	db1.Close()

	db2, _ := Open(path)
	defer db2.Close()

	entries, err := db2.PendingList()
	if err != nil {
		t.Fatalf("PendingList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after reopen, got %d", len(entries))
	}
	if entries[0].URL != "https://example.com/persist" {
		t.Errorf("URL = %q, want %q", entries[0].URL, "https://example.com/persist")
	}
}
