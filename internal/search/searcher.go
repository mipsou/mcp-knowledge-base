/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package search

// Result represents a single search result.
type Result struct {
	Corpus  string  `json:"corpus"`
	DocName string  `json:"doc_name"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
}

// Searcher is the interface for all search backends.
type Searcher interface {
	// Index processes a document and adds it to the search index.
	Index(corpus, docName, content string) error

	// Search returns ranked results for a query.
	Search(query string, maxResults int) ([]Result, error)

	// Remove removes a document from the index.
	Remove(corpus, docName string) error
}
