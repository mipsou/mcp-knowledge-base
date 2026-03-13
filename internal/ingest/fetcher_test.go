/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package ingest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchConvertsHTMLToMarkdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body><h1>Hello</h1><p>World</p></body></html>`))
	}))
	defer srv.Close()

	f := newTestFetcher()
	md, err := f.Fetch(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md == "" {
		t.Fatal("expected non-empty markdown")
	}
	if !strings.Contains(md, "Hello") {
		t.Errorf("expected markdown to contain 'Hello', got %q", md)
	}
	if !strings.Contains(md, "World") {
		t.Errorf("expected markdown to contain 'World', got %q", md)
	}
}

func TestFetchRejectsInvalidURL(t *testing.T) {
	f := NewFetcher()
	_, err := f.Fetch("not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestFetchRejectsNonHTTP(t *testing.T) {
	f := NewFetcher()
	_, err := f.Fetch("ftp://example.com/file")
	if err == nil {
		t.Fatal("expected error for non-HTTP scheme")
	}
}

func TestFetchRejectsPrivateIP(t *testing.T) {
	f := NewFetcher()
	// SSRF protection: block requests to private IPs.
	_, err := f.Fetch("http://127.0.0.1:1234/test")
	if err == nil {
		t.Fatal("expected error for private IP (SSRF protection)")
	}
}

func TestFetchHandles404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	f := newTestFetcher()
	_, err := f.Fetch(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
