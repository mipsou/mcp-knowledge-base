/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package search

import (
	"testing"
)

func TestBM25IndexAndSearch(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "caddy.md", "Caddy is a web server with automatic HTTPS")
	_ = b.Index("infra", "nginx.md", "Nginx is a high performance web server")

	results, err := b.Search("caddy HTTPS", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	if results[0].DocName != "caddy.md" {
		t.Errorf("expected caddy.md first, got %q", results[0].DocName)
	}
}

func TestBM25SearchNoResults(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "caddy.md", "Caddy is a web server")

	results, err := b.Search("kubernetes deployment", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestBM25SearchMaxResults(t *testing.T) {
	b := NewBM25()
	for i := 0; i < 20; i++ {
		_ = b.Index("infra", "doc"+string(rune('a'+i))+".md", "server configuration guide")
	}

	results, err := b.Search("server", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) > 5 {
		t.Errorf("expected max 5 results, got %d", len(results))
	}
}

func TestBM25RankingQuality(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "podman.md", "Podman is a container engine. Podman runs containers without a daemon. Podman is rootless.")
	_ = b.Index("infra", "docker.md", "Docker is a container platform. It requires a daemon.")
	_ = b.Index("infra", "caddy.md", "Caddy is a web server with automatic HTTPS.")

	results, err := b.Search("podman container rootless", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	if results[0].DocName != "podman.md" {
		t.Errorf("expected podman.md first, got %q", results[0].DocName)
	}
}

func TestBM25Remove(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "caddy.md", "Caddy web server")
	_ = b.Remove("infra", "caddy.md")

	results, err := b.Search("caddy", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results after remove, got %d", len(results))
	}
}

func TestBM25SearchAcrossCorpora(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "caddy.md", "Caddy web server HTTPS")
	_ = b.Index("podman", "basics.md", "Podman container basics")

	results, err := b.Search("server", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected cross-corpus results")
	}
}

func TestBM25EmptyQuery(t *testing.T) {
	b := NewBM25()
	_ = b.Index("infra", "caddy.md", "Caddy web server")

	results, err := b.Search("", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for empty query, got %d", len(results))
	}
}
