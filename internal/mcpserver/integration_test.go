/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package mcpserver

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

// TestIntegrationFullWorkflow tests the complete workflow:
// create corpus → add documents → search → read document.
func TestIntegrationFullWorkflow(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	bm25 := search.NewBM25()
	db, err := storage.Open(filepath.Join(root, "biblium.db"))
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer db.Close()
	s := New(store, bm25, db)

	// 1. Create a corpus.
	text := callTool(t, s, "create_corpus", map[string]any{"name": "homelab"})
	if !strings.Contains(text, "homelab") {
		t.Fatalf("create_corpus: %s", text)
	}

	// 2. Add documents.
	callTool(t, s, "add_document", map[string]any{
		"corpus":  "homelab",
		"name":    "caddy.md",
		"content": "Caddy is a modern web server with automatic HTTPS using Let's Encrypt certificates",
	})
	callTool(t, s, "add_document", map[string]any{
		"corpus":  "homelab",
		"name":    "pihole.md",
		"content": "Pi-hole is a DNS sinkhole that blocks ads and trackers at the network level",
	})
	callTool(t, s, "add_document", map[string]any{
		"corpus":  "homelab",
		"name":    "unbound.md",
		"content": "Unbound is a recursive DNS resolver that provides privacy and validation",
	})

	// 3. List corpus entries.
	text = callTool(t, s, "list_corpus", nil)
	if !strings.Contains(text, "homelab") {
		t.Fatalf("list_corpus: %s", text)
	}

	// 4. List documents.
	text = callTool(t, s, "list_documents", map[string]any{"corpus": "homelab"})
	if !strings.Contains(text, "caddy.md") || !strings.Contains(text, "pihole.md") {
		t.Fatalf("list_documents: %s", text)
	}

	// 5. Search.
	text = callTool(t, s, "search", map[string]any{
		"query":       "DNS resolver",
		"max_results": 3,
	})
	if !strings.Contains(text, "unbound.md") {
		t.Errorf("search for 'DNS resolver' should find unbound.md, got: %s", text)
	}

	// 6. Read a specific document.
	text = callTool(t, s, "read_document", map[string]any{
		"corpus": "homelab",
		"name":   "caddy.md",
	})
	if !strings.Contains(text, "Caddy") || !strings.Contains(text, "HTTPS") {
		t.Errorf("read_document caddy.md unexpected content: %s", text)
	}

	// 7. Suggest URL (pending workflow).
	text = callTool(t, s, "suggest_url", map[string]any{
		"corpus": "homelab",
		"url":    "https://example.com/docs/coreos",
	})
	if !strings.Contains(text, "pending") {
		t.Errorf("suggest_url should return pending status, got: %s", text)
	}

	// 8. List pending.
	text = callTool(t, s, "list_pending", nil)
	if !strings.Contains(text, "example.com") {
		t.Errorf("list_pending should show the suggested URL, got: %s", text)
	}
}

// TestPersistenceSurvivesRestart simulates a server restart:
// add docs → "restart" (new BM25 rebuilt from disk) → search still works.
func TestPersistenceSurvivesRestart(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	dbPath := filepath.Join(root, "biblium.db")

	// --- Session 1: add documents + suggest URL ---
	bm25 := search.NewBM25()
	db1, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	s1 := New(store, bm25, db1)

	callTool(t, s1, "create_corpus", map[string]any{"name": "infra"})
	callTool(t, s1, "add_document", map[string]any{
		"corpus":  "infra",
		"name":    "caddy.md",
		"content": "Caddy is a web server with automatic HTTPS",
	})
	callTool(t, s1, "add_document", map[string]any{
		"corpus":  "infra",
		"name":    "pihole.md",
		"content": "Pi-hole blocks ads at DNS level",
	})
	callTool(t, s1, "suggest_url", map[string]any{
		"corpus": "infra",
		"url":    "https://example.com/persist-test",
	})
	db1.Close()

	// --- Session 2: simulate restart — new BM25 rebuilt from Walk ---
	bm25v2 := search.NewBM25()
	var indexed int
	err = store.Walk(func(c, name, content string) error {
		indexed++
		return bm25v2.Index(c, name, content)
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if indexed != 2 {
		t.Fatalf("expected 2 docs indexed, got %d", indexed)
	}

	db2, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	defer db2.Close()
	s2 := New(store, bm25v2, db2)

	// Search should still work after "restart".
	text := callTool(t, s2, "search", map[string]any{
		"query":       "HTTPS web server",
		"max_results": 5,
	})
	if !strings.Contains(text, "caddy.md") {
		t.Errorf("search after restart should find caddy.md, got: %s", text)
	}

	// Pending URLs should survive restart (SQLite).
	text = callTool(t, s2, "list_pending", nil)
	if !strings.Contains(text, "persist-test") {
		t.Errorf("pending URLs should survive restart, got: %s", text)
	}
}
