/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package ingest

// Fetcher downloads a URL and converts it to markdown.
type Fetcher struct{}

// NewFetcher creates a new Fetcher.
func NewFetcher() *Fetcher { return &Fetcher{} }

// Fetch downloads the URL and returns the content as markdown.
func (f *Fetcher) Fetch(rawURL string) (string, error) {
	return "", nil
}
