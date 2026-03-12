/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

func openDB(t *testing.T, root string) *storage.DB {
	t.Helper()
	db, err := storage.Open(filepath.Join(root, "biblium.db"))
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// callTool sends a tools/call JSON-RPC request and returns the text content.
func callTool(t *testing.T, s *Server, toolName string, args map[string]any) string {
	t.Helper()

	params := map[string]any{
		"name":      toolName,
		"arguments": args,
	}
	paramsBytes, _ := json.Marshal(params)

	reqJSON := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + string(paramsBytes) + `}`)

	ctx := context.Background()
	result := s.MCPServer().HandleMessage(ctx, reqJSON)

	respBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("JSON-RPC error: %s", resp.Error.Message)
	}

	if resp.Result.IsError {
		if len(resp.Result.Content) > 0 {
			return "ERROR:" + resp.Result.Content[0].Text
		}
		return "ERROR:unknown"
	}

	if len(resp.Result.Content) == 0 {
		return ""
	}
	return resp.Result.Content[0].Text
}

func TestHandlerCreateCorpus(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "create_corpus", map[string]any{"name": "infra"})
	if !strings.Contains(text, "infra") {
		t.Errorf("expected response to mention corpus name, got %q", text)
	}
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}

	// Verify corpus actually exists on disk.
	entries, err := store.List()
	if err != nil {
		t.Fatalf("store.List error: %v", err)
	}
	found := false
	for _, c := range entries {
		if c == "infra" {
			found = true
		}
	}
	if !found {
		t.Error("corpus 'infra' not found in store after create_corpus")
	}
}

func TestHandlerListCorpus(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	_ = store.Create("alpha")
	_ = store.Create("beta")
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "list_corpus", nil)
	if !strings.Contains(text, "alpha") || !strings.Contains(text, "beta") {
		t.Errorf("expected both entries in response, got %q", text)
	}
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}
}

func TestHandlerAddDocument(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	_ = store.Create("infra")
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "add_document", map[string]any{
		"corpus":  "infra",
		"name":    "caddy.md",
		"content": "Caddy is a web server",
	})
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}

	// Verify document exists.
	data, err := store.ReadDoc("infra", "caddy.md")
	if err != nil {
		t.Fatalf("ReadDoc error: %v", err)
	}
	if string(data) != "Caddy is a web server" {
		t.Errorf("document content = %q, want %q", string(data), "Caddy is a web server")
	}
}

func TestHandlerListDocuments(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	_ = store.Create("infra")
	_ = store.AddDoc("infra", "caddy.md", []byte("Caddy"))
	_ = store.AddDoc("infra", "nginx.md", []byte("Nginx"))
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "list_documents", map[string]any{"corpus": "infra"})
	if !strings.Contains(text, "caddy.md") || !strings.Contains(text, "nginx.md") {
		t.Errorf("expected both docs in response, got %q", text)
	}
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}
}

func TestHandlerReadDocument(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	_ = store.Create("infra")
	_ = store.AddDoc("infra", "caddy.md", []byte("Caddy is great"))
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "read_document", map[string]any{
		"corpus": "infra",
		"name":   "caddy.md",
	})
	if !strings.Contains(text, "Caddy is great") {
		t.Errorf("expected document content in response, got %q", text)
	}
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}
}

func TestHandlerSearch(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	_ = store.Create("infra")
	bm25 := search.NewBM25()
	_ = bm25.Index("infra", "caddy.md", "Caddy is a web server with HTTPS")
	s := New(store, bm25, openDB(t, root))

	text := callTool(t, s, "search", map[string]any{
		"query":       "caddy HTTPS",
		"max_results": 5,
	})
	if !strings.Contains(text, "caddy.md") {
		t.Errorf("expected caddy.md in search results, got %q", text)
	}
	if strings.HasPrefix(text, "not implemented") {
		t.Error("handler still returns stub response")
	}
}

func TestHandlerCreateCorpusTraversal(t *testing.T) {
	root := t.TempDir()
	store := corpus.NewFileStore(root)
	s := New(store, search.NewBM25(), openDB(t, root))

	text := callTool(t, s, "create_corpus", map[string]any{"name": "../escape"})
	if !strings.Contains(text, "ERROR:") && !strings.Contains(strings.ToLower(text), "error") {
		t.Errorf("expected error for path traversal, got %q", text)
	}
}
