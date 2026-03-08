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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/lore-mcp/internal/corpus"
	"github.com/mipsou/lore-mcp/internal/search"
)

// Server wraps the MCP server with Lore-specific configuration.
type Server struct {
	mcp    *server.MCPServer
	store  *corpus.FileStore
	search search.Searcher
}

// New creates a new Lore MCP server with all tools registered.
func New(store *corpus.FileStore, searcher search.Searcher) *Server {
	s := server.NewMCPServer(
		"lore",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	srv := &Server{
		mcp:    s,
		store:  store,
		search: searcher,
	}
	srv.registerTools()

	return srv
}

// MCPServer returns the underlying mcp-go server for wiring.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcp
}

// registerTools registers all MCP tools on the server.
func (s *Server) registerTools() {
	s.mcp.AddTool(
		mcp.NewTool("create_corpus",
			mcp.WithDescription("Create a new knowledge corpus"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Corpus name")),
		),
		s.stubHandler,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_corpora",
			mcp.WithDescription("List all available corpora"),
		),
		s.stubHandler,
	)

	s.mcp.AddTool(
		mcp.NewTool("add_document",
			mcp.WithDescription("Add a document to a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Target corpus name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Document content")),
		),
		s.stubHandler,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_documents",
			mcp.WithDescription("List documents in a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Corpus name")),
		),
		s.stubHandler,
	)

	s.mcp.AddTool(
		mcp.NewTool("read_document",
			mcp.WithDescription("Read a document from a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Corpus name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
		),
		s.stubHandler,
	)

	s.mcp.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("Search across corpora"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithNumber("max_results", mcp.Description("Maximum results to return")),
		),
		s.stubHandler,
	)
}

// stubHandler is a placeholder handler that returns "not implemented".
func (s *Server) stubHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("not implemented"), nil
}
