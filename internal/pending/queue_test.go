/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package pending

import (
	"testing"
)

func TestAddReturnsEntry(t *testing.T) {
	q := NewQueue()
	entry, err := q.Add("infra", "https://example.com/doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry")
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
	if entry.Status != StatusPending {
		t.Errorf("Status = %q, want %q", entry.Status, StatusPending)
	}
}

func TestListReturnsPendingEntries(t *testing.T) {
	q := NewQueue()
	_, _ = q.Add("infra", "https://example.com/a")
	_, _ = q.Add("infra", "https://example.com/b")

	entries := q.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestApproveChangesStatus(t *testing.T) {
	q := NewQueue()
	entry, _ := q.Add("infra", "https://example.com/doc")

	approved, err := q.Approve(entry.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved == nil {
		t.Fatal("expected non-nil entry")
	}
	if approved.Status != StatusApproved {
		t.Errorf("Status = %q, want %q", approved.Status, StatusApproved)
	}
}

func TestRejectChangesStatus(t *testing.T) {
	q := NewQueue()
	entry, _ := q.Add("infra", "https://example.com/doc")

	rejected, err := q.Reject(entry.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected == nil {
		t.Fatal("expected non-nil entry")
	}
	if rejected.Status != StatusRejected {
		t.Errorf("Status = %q, want %q", rejected.Status, StatusRejected)
	}
}

func TestApproveUnknownIDReturnsError(t *testing.T) {
	q := NewQueue()
	_, err := q.Approve("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
}

func TestRejectUnknownIDReturnsError(t *testing.T) {
	q := NewQueue()
	_, err := q.Reject("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
}
