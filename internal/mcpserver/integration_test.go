/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package mcpserver

import (
	"strings"
	"testing"

	"github.com/mipsou/lore-mcp/internal/corpus"
	"github.com/mipsou/lore-mcp/internal/search"
)

// TestIntegrationFullWorkflow tests the complete workflow:
// create corpus → add documents → search → read document.
func TestIntegrationFullWorkflow(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	bm25 := search.NewBM25()
	s := New(store, bm25)

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

	// 3. List corpora.
	text = callTool(t, s, "list_corpora", nil)
	if !strings.Contains(text, "homelab") {
		t.Fatalf("list_corpora: %s", text)
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
