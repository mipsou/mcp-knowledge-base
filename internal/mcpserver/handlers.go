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
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleCreateCorpus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing required parameter: name")), nil
	}

	if err := s.store.Create(name); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error creating corpus: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("corpus %q created", name)), nil
}

func (s *Server) handleListCorpora(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	corpora, err := s.store.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error listing corpora: %v", err)), nil
	}

	if len(corpora) == 0 {
		return mcp.NewToolResultText("no corpora found"), nil
	}

	return mcp.NewToolResultText(strings.Join(corpora, "\n")), nil
}

func (s *Server) handleAddDocument(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	corpusName, err := req.RequireString("corpus")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: corpus"), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: name"), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: content"), nil
	}

	if err := s.store.AddDoc(corpusName, name, []byte(content)); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error adding document: %v", err)), nil
	}

	// Index the document for search.
	if err := s.search.Index(corpusName, name, content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error indexing document: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("document %q added to corpus %q", name, corpusName)), nil
}

func (s *Server) handleListDocuments(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	corpusName, err := req.RequireString("corpus")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: corpus"), nil
	}

	docs, err := s.store.ListDocs(corpusName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error listing documents: %v", err)), nil
	}

	if len(docs) == 0 {
		return mcp.NewToolResultText("no documents found"), nil
	}

	return mcp.NewToolResultText(strings.Join(docs, "\n")), nil
}

func (s *Server) handleReadDocument(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	corpusName, err := req.RequireString("corpus")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: corpus"), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: name"), nil
	}

	data, err := s.store.ReadDoc(corpusName, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error reading document: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleSearch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: query"), nil
	}
	maxResults := req.GetInt("max_results", 10)

	results, err := s.search.Search(query, maxResults)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search error: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("no results found"), nil
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error formatting results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(out)), nil
}
