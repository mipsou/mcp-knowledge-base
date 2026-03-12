/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/ingest"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

// Server wraps the MCP server with Biblium-specific configuration.
type Server struct {
	mcp     *server.MCPServer
	store   *corpus.FileStore
	search  search.Searcher
	db      *storage.DB
	fetcher *ingest.Fetcher
}

// New creates a new Biblium MCP server with all tools registered.
func New(store *corpus.FileStore, searcher search.Searcher, db *storage.DB) *Server {
	s := server.NewMCPServer(
		"biblium",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	srv := &Server{
		mcp:     s,
		store:   store,
		search:  searcher,
		db:      db,
		fetcher: ingest.NewFetcher(),
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
		s.handleCreateCorpus,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_corpus",
			mcp.WithDescription("List all available corpus entries"),
		),
		s.handleListCorpus,
	)

	s.mcp.AddTool(
		mcp.NewTool("add_document",
			mcp.WithDescription("Add a document to a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Target corpus name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Document content")),
		),
		s.handleAddDocument,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_documents",
			mcp.WithDescription("List documents in a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Corpus name")),
		),
		s.handleListDocuments,
	)

	s.mcp.AddTool(
		mcp.NewTool("read_document",
			mcp.WithDescription("Read a document from a corpus"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Corpus name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
		),
		s.handleReadDocument,
	)

	s.mcp.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("Search across all corpus entries"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithNumber("max_results", mcp.Description("Maximum results to return")),
		),
		s.handleSearch,
	)

	s.mcp.AddTool(
		mcp.NewTool("suggest_url",
			mcp.WithDescription("Suggest a URL for ingestion into a corpus (requires approval)"),
			mcp.WithString("corpus", mcp.Required(), mcp.Description("Target corpus name")),
			mcp.WithString("url", mcp.Required(), mcp.Description("URL to ingest")),
		),
		s.handleSuggestURL,
	)

	s.mcp.AddTool(
		mcp.NewTool("approve_url",
			mcp.WithDescription("Approve a pending URL for ingestion"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Pending entry ID")),
		),
		s.handleApproveURL,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_pending",
			mcp.WithDescription("List all pending URL suggestions"),
		),
		s.handleListPending,
	)
}
