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

	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-biblium/internal/config"
	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/mcpserver"
	"github.com/mipsou/mcp-biblium/internal/search"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "biblium: config error: %v\n", err)
		os.Exit(1)
	}

	store := corpus.NewFileStore(cfg.DataDir)
	searcher := search.NewBM25()
	srv := mcpserver.New(store, searcher)

	fmt.Fprintf(os.Stderr, "biblium: starting (data=%s, search=%s)\n", cfg.DataDir, cfg.SearchBackend)

	if err := server.ServeStdio(srv.MCPServer()); err != nil {
		fmt.Fprintf(os.Stderr, "biblium: server error: %v\n", err)
		os.Exit(1)
	}
}
