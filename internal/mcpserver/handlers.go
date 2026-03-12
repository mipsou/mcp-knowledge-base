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

func (s *Server) handleSuggestURL(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	corpusName, err := req.RequireString("corpus")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: corpus"), nil
	}
	rawURL, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: url"), nil
	}

	entry, err := s.pending.Add(corpusName, rawURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error adding to pending: %v", err)), nil
	}

	out, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error formatting entry: %v", err)), nil
	}

	return mcp.NewToolResultText(string(out)), nil
}

func (s *Server) handleApproveURL(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: id"), nil
	}

	entry, err := s.pending.Approve(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error approving: %v", err)), nil
	}

	// Fetch the URL and ingest into the corpus.
	md, err := s.fetcher.Fetch(entry.URL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error fetching URL: %v", err)), nil
	}

	// Generate a document name from the URL.
	docName := sanitizeDocName(entry.URL)

	if err := s.store.AddDoc(entry.Corpus, docName, []byte(md)); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error storing document: %v", err)), nil
	}

	if err := s.search.Index(entry.Corpus, docName, md); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error indexing document: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("approved and ingested %q as %q in corpus %q", entry.URL, docName, entry.Corpus)), nil
}

func (s *Server) handleListPending(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entries := s.pending.List()
	if len(entries) == 0 {
		return mcp.NewToolResultText("no pending entries"), nil
	}

	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error formatting entries: %v", err)), nil
	}

	return mcp.NewToolResultText(string(out)), nil
}

// sanitizeDocName converts a URL to a safe document filename.
func sanitizeDocName(rawURL string) string {
	name := rawURL
	// Remove scheme.
	for _, prefix := range []string{"https://", "http://"} {
		name = strings.TrimPrefix(name, prefix)
	}
	// Replace unsafe characters.
	replacer := strings.NewReplacer("/", "_", "?", "_", "&", "_", "=", "_", "#", "_", ":", "_")
	name = replacer.Replace(name)
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	return name
}
