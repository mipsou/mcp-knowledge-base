/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package mcpserver

import (
	"context"
	"encoding/json"
	"testing"

	"path/filepath"

	"github.com/mipsou/mcp-biblium/internal/filestore"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	root := t.TempDir()
	store := filestore.New(root)
	searcher := search.NewBM25()
	db, err := storage.Open(filepath.Join(root, "biblium.db"))
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(store, searcher, db, "test")
}

func TestNewServerNotNil(t *testing.T) {
	s := newTestServer(t)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.MCPServer() == nil {
		t.Fatal("expected non-nil underlying MCP server")
	}
}

func TestServerRegistersTools(t *testing.T) {
	s := newTestServer(t)

	// Build a JSON-RPC request for tools/list.
	reqJSON := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)

	ctx := context.Background()
	result := s.MCPServer().HandleMessage(ctx, reqJSON)

	// Marshal the response to inspect it.
	respBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var resp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Expect all 9 tools to be registered.
	expectedTools := []string{
		"create_collection",
		"list_collections",
		"add_document",
		"list_documents",
		"read_document",
		"search",
		"suggest_url",
		"approve_url",
		"list_pending",
	}
	toolMap := make(map[string]bool)
	for _, tool := range resp.Result.Tools {
		toolMap[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !toolMap[name] {
			t.Errorf("expected tool %q to be registered, but it was not found", name)
		}
	}
}
