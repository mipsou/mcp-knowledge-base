/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-biblium/internal/filestore"
	"github.com/mipsou/mcp-biblium/internal/ingest"
	"github.com/mipsou/mcp-biblium/internal/search"
	"github.com/mipsou/mcp-biblium/internal/storage"
)

// Server wraps the MCP server with Biblium-specific configuration.
type Server struct {
	mcp     *server.MCPServer
	store   *filestore.Store
	search  search.Searcher
	db      *storage.DB
	fetcher *ingest.Fetcher
}

// New creates a new Biblium MCP server with all tools registered.
func New(store *filestore.Store, searcher search.Searcher, db *storage.DB, version string) *Server {
	s := server.NewMCPServer(
		"biblium",
		version,
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
		mcp.NewTool("create_collection",
			mcp.WithDescription("Create a new knowledge collection"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Collection name")),
		),
		s.handleCreateCollection,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_collections",
			mcp.WithDescription("List all available collections"),
		),
		s.handleListCollections,
	)

	s.mcp.AddTool(
		mcp.NewTool("add_document",
			mcp.WithDescription("Add a document to a collection"),
			mcp.WithString("collection", mcp.Required(), mcp.Description("Target collection name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Document content")),
		),
		s.handleAddDocument,
	)

	s.mcp.AddTool(
		mcp.NewTool("list_documents",
			mcp.WithDescription("List documents in a collection"),
			mcp.WithString("collection", mcp.Required(), mcp.Description("Collection name")),
		),
		s.handleListDocuments,
	)

	s.mcp.AddTool(
		mcp.NewTool("read_document",
			mcp.WithDescription("Read a document from a collection"),
			mcp.WithString("collection", mcp.Required(), mcp.Description("Collection name")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Document filename")),
		),
		s.handleReadDocument,
	)

	s.mcp.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("Search across all collections"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithNumber("max_results", mcp.Description("Maximum results to return")),
		),
		s.handleSearch,
	)

	s.mcp.AddTool(
		mcp.NewTool("suggest_url",
			mcp.WithDescription("Suggest a URL for ingestion into a collection (requires approval)"),
			mcp.WithString("collection", mcp.Required(), mcp.Description("Target collection name")),
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
