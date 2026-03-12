/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all runtime configuration.
type Config struct {
	DataDir       string
	SearchBackend string
	LogLevel      string
	OllamaURL     string
	OllamaModel   string
}

var validBackends = map[string]bool{
	"bm25":   true,
	"ollama": true,
}

// Load reads configuration from environment variables with defaults.
func Load() (*Config, error) {
	cfg := &Config{
		DataDir:       envOr("LORE_DATA_DIR", defaultDataDir()),
		SearchBackend: envOr("LORE_SEARCH_BACKEND", "bm25"),
		LogLevel:      envOr("LORE_LOG_LEVEL", "info"),
		OllamaURL:     envOr("LORE_OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:   envOr("LORE_OLLAMA_MODEL", "all-minilm:l6-v2"),
	}

	if !validBackends[cfg.SearchBackend] {
		return nil, fmt.Errorf("config: invalid search backend %q (valid: bm25, ollama)", cfg.SearchBackend)
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "lore_data"
	}
	return filepath.Join(home, "lore_data")
}
