/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package search

import (
	"math"
	"sort"
	"strings"
	"sync"
)

// BM25 parameters (Okapi BM25 defaults).
const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

// docKey uniquely identifies a document across corpora.
type docKey struct {
	corpus  string
	docName string
}

// docEntry stores indexed data for a single document.
type docEntry struct {
	key     docKey
	content string
	terms   map[string]int // term -> frequency
	length  int            // total term count
}

// BM25 implements the Searcher interface using the BM25 ranking algorithm.
// Pure Go, zero external dependencies.
type BM25 struct {
	mu   sync.RWMutex
	docs map[docKey]*docEntry
	df   map[string]int // term -> number of documents containing term
}

// NewBM25 creates a new BM25 search engine.
func NewBM25() *BM25 {
	return &BM25{
		docs: make(map[docKey]*docEntry),
		df:   make(map[string]int),
	}
}

// tokenize splits text into lowercase terms.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Split on any non-alphanumeric character.
	var tokens []string
	var current strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// termFrequencies counts occurrences of each term.
func termFrequencies(tokens []string) map[string]int {
	freq := make(map[string]int, len(tokens))
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}

// Index processes a document and adds it to the search index.
func (b *BM25) Index(corpus, docName, content string) error {
	key := docKey{corpus: corpus, docName: docName}
	tokens := tokenize(content)
	tf := termFrequencies(tokens)

	b.mu.Lock()
	defer b.mu.Unlock()

	// If document already exists, remove old DF contributions.
	if old, exists := b.docs[key]; exists {
		for term := range old.terms {
			b.df[term]--
			if b.df[term] <= 0 {
				delete(b.df, term)
			}
		}
	}

	entry := &docEntry{
		key:     key,
		content: content,
		terms:   tf,
		length:  len(tokens),
	}
	b.docs[key] = entry

	// Update document frequency.
	for term := range tf {
		b.df[term]++
	}

	return nil
}

// Remove removes a document from the index.
func (b *BM25) Remove(corpus, docName string) error {
	key := docKey{corpus: corpus, docName: docName}

	b.mu.Lock()
	defer b.mu.Unlock()

	old, exists := b.docs[key]
	if !exists {
		return nil
	}

	for term := range old.terms {
		b.df[term]--
		if b.df[term] <= 0 {
			delete(b.df, term)
		}
	}
	delete(b.docs, key)

	return nil
}

// Search returns ranked results for a query.
func (b *BM25) Search(query string, maxResults int) ([]Result, error) {
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil, nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	n := len(b.docs)
	if n == 0 {
		return nil, nil
	}

	// Compute average document length.
	var totalLen int
	for _, doc := range b.docs {
		totalLen += doc.length
	}
	avgDL := float64(totalLen) / float64(n)

	// Score each document.
	type scored struct {
		key   docKey
		score float64
	}
	var results []scored

	for _, doc := range b.docs {
		score := 0.0
		for _, term := range queryTerms {
			tf, ok := doc.terms[term]
			if !ok {
				continue
			}
			dfVal := b.df[term]
			// IDF: log((N - df + 0.5) / (df + 0.5) + 1)
			idf := math.Log((float64(n)-float64(dfVal)+0.5)/(float64(dfVal)+0.5) + 1.0)
			// TF component: (tf * (k1 + 1)) / (tf + k1 * (1 - b + b * dl/avgdl))
			tfNorm := (float64(tf) * (bm25K1 + 1.0)) /
				(float64(tf) + bm25K1*(1.0-bm25B+bm25B*float64(doc.length)/avgDL))
			score += idf * tfNorm
		}
		if score > 0 {
			results = append(results, scored{key: doc.key, score: score})
		}
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Limit results.
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	// Build output.
	out := make([]Result, len(results))
	for i, r := range results {
		doc := b.docs[r.key]
		snippet := doc.content
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		out[i] = Result{
			Corpus:  r.key.corpus,
			DocName: r.key.docName,
			Score:   r.score,
			Snippet: snippet,
		}
	}

	return out, nil
}
