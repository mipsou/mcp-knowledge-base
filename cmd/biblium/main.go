/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-biblium/internal/config"
	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/mcpserver"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "biblium: config error: %v\n", err)
		os.Exit(1)
	}

	// Ensure data directory exists.
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "biblium: mkdir error: %v\n", err)
		os.Exit(1)
	}

	// Open SQLite database.
	db, err := storage.Open(filepath.Join(cfg.DataDir, "biblium.db"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "biblium: db error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store := corpus.NewFileStore(cfg.DataDir)
	searcher := search.NewBM25()

	// Rebuild BM25 index from existing documents on disk.
	var indexed int
	err = store.Walk(func(c, name, content string) error {
		indexed++
		return searcher.Index(c, name, content)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "biblium: index rebuild error: %v\n", err)
		os.Exit(1)
	}

	srv := mcpserver.New(store, searcher, db)

	fmt.Fprintf(os.Stderr, "biblium: starting (data=%s, search=%s, indexed=%d docs)\n",
		cfg.DataDir, cfg.SearchBackend, indexed)

	if err := server.ServeStdio(srv.MCPServer()); err != nil {
		fmt.Fprintf(os.Stderr, "biblium: server error: %v\n", err)
		os.Exit(1)
	}
}
