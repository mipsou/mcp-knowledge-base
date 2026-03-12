/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty")
	}
	if cfg.SearchBackend != "bm25" {
		t.Errorf("SearchBackend got %q, want %q", cfg.SearchBackend, "bm25")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel got %q, want %q", cfg.LogLevel, "info")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("LORE_DATA_DIR", "/tmp/lore-test")
	t.Setenv("LORE_SEARCH_BACKEND", "ollama")
	t.Setenv("LORE_LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDir != "/tmp/lore-test" {
		t.Errorf("DataDir got %q, want %q", cfg.DataDir, "/tmp/lore-test")
	}
	if cfg.SearchBackend != "ollama" {
		t.Errorf("SearchBackend got %q, want %q", cfg.SearchBackend, "ollama")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel got %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestLoadRejectsInvalidBackend(t *testing.T) {
	t.Setenv("LORE_SEARCH_BACKEND", "invalid")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid backend, got nil")
	}
}

func TestLoadOllamaDefaults(t *testing.T) {
	t.Setenv("LORE_SEARCH_BACKEND", "ollama")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OllamaURL == "" {
		t.Error("OllamaURL should have a default")
	}
	if cfg.OllamaModel == "" {
		t.Error("OllamaModel should have a default")
	}
}
