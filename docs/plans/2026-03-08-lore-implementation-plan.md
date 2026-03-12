# Biblium MCP Server — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a pure-Go MCP knowledge base server (binary name: `biblium`) with BM25 search, URL ingestion, and pluggable search backends — under 3000 lines.

**Architecture:** Single Go binary, stdio MCP transport via `mark3labs/mcp-go`, file-based corpus storage (markdown docs + JSON index), `Searcher` interface with BM25 default and Ollama optional. Security-first: SafePath for all file ops, URL validation for ingestion.

**Tech Stack:** Go 1.21+, `mark3labs/mcp-go` (MCP SDK), `JohannesKaufmann/html-to-markdown` (HTML→MD), stdlib for everything else.

**Dependencies (3 total):**
- `github.com/mark3labs/mcp-go` — MCP protocol (MIT)
- `github.com/JohannesKaufmann/html-to-markdown/v2` — HTML conversion (MIT)
- `github.com/google/uuid` — pending URL IDs (BSD-3-Clause)

---

## Task 1: Create branch and init Go module

**Files:**
- Create: `go.mod`
- Create: `LICENSE`
- Create: `.gitignore`

**Step 1: Create dev-rewrite branch**

```bash
git checkout main
git checkout -b dev-rewrite
```

**Step 2: Init Go module**

```bash
go mod init github.com/mipsou/mcp-biblium
```

**Step 3: Create LICENSE (EUPL-1.2)**

Copy the full EUPL-1.2 text into `LICENSE`.
First line: `EUROPEAN UNION PUBLIC LICENCE v. 1.2`

**Step 4: Create .gitignore**

```gitignore
# Binary
biblium
biblium.exe

# Build
/build/
/dist/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Test
coverage.out
coverage.html

# Secrets
.env
*.key
*.pem
```

**Step 5: Commit**

```bash
git add go.mod LICENSE .gitignore
git commit -m "chore: init Go module with EUPL-1.2 license"
```

---

## Task 2: Create main.go entry point

**Files:**
- Create: `cmd/biblium/main.go`

**Step 1: Write main.go**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "biblium: starting")
	os.Exit(0)
}
```

**Step 2: Verify it compiles and runs**

Run: `go build -o biblium ./cmd/biblium/ && ./biblium`
Expected: prints "biblium: starting" to stderr, exits 0

**Step 3: Commit**

```bash
git add cmd/biblium/main.go
git commit -m "chore: add minimal main.go entry point"
```

---

## Task 3: SafePath — write failing tests

**Files:**
- Create: `internal/safepath/safepath.go` (types only)
- Create: `internal/safepath/safepath_test.go`

**Step 1: Write the test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package safepath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveValidPath(t *testing.T) {
	root := t.TempDir()
	got, err := Resolve(root, "subdir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "subdir")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveNestedPath(t *testing.T) {
	root := t.TempDir()
	got, err := Resolve(root, "a/b/c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "a", "b", "c")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveBlocksTraversalDotDot(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, "../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
}

func TestResolveBlocksTraversalHidden(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, "subdir/../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for hidden traversal, got nil")
	}
}

func TestResolveBlocksAbsolutePath(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, "/etc/passwd")
	if err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
}

func TestResolveBlocksEmptyName(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, "")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestResolveBlocksDotOnly(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(root, ".")
	if err == nil {
		t.Fatal("expected error for dot-only path, got nil")
	}
}

func TestResolveBlocksSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	// Create a symlink inside root that points outside
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	err := os.Symlink(outside, link)
	if err != nil {
		t.Skip("symlinks not supported on this OS")
	}
	_, err = Resolve(root, "escape/file.txt")
	if err == nil {
		t.Fatal("expected error for symlink escape, got nil")
	}
}
```

**Step 2: Write minimal types so the file compiles**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package safepath

// Resolve validates and resolves a user-provided path segment
// against a trusted root directory.
// Returns the absolute resolved path or an error if the path
// escapes the root.
func Resolve(root, userPath string) (string, error) {
	return "", nil // TDD: stub
}
```

**Step 3: Run tests to verify they FAIL**

Run: `go test ./internal/safepath/ -v`
Expected: Most tests FAIL (stub returns empty string / no error)

**Step 4: Commit**

```bash
git add internal/safepath/
git commit -m "test: add SafePath tests (RED)"
```

---

## Task 4: SafePath — implement to make tests green

**Files:**
- Modify: `internal/safepath/safepath.go`

**Step 1: Implement Resolve**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package safepath

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrTraversal    = errors.New("safepath: path traversal detected")
	ErrEmptyPath    = errors.New("safepath: empty path")
	ErrAbsolutePath = errors.New("safepath: absolute path not allowed")
)

// Resolve validates and resolves a user-provided path segment
// against a trusted root directory.
// Returns the absolute resolved path or an error if the path
// escapes the root.
func Resolve(root, userPath string) (string, error) {
	if userPath == "" {
		return "", ErrEmptyPath
	}

	// Block absolute paths
	if filepath.IsAbs(userPath) {
		return "", ErrAbsolutePath
	}

	// Clean the path to resolve . and ..
	cleaned := filepath.Clean(userPath)

	// Block paths that resolve to "." (root itself)
	if cleaned == "." {
		return "", ErrEmptyPath
	}

	// Block paths that start with ".."
	if strings.HasPrefix(cleaned, "..") {
		return "", ErrTraversal
	}

	// Resolve root to absolute
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	// Join and resolve the full path
	joined := filepath.Join(absRoot, cleaned)

	// Evaluate symlinks to detect escapes
	resolved, err := filepath.EvalSymlinks(filepath.Dir(joined))
	if err != nil {
		// If parent dir doesn't exist yet, verify without symlink eval
		// by checking the cleaned path doesn't escape
		resolved = joined
	} else {
		resolved = filepath.Join(resolved, filepath.Base(joined))
	}

	// Verify the resolved path is under root
	if !strings.HasPrefix(resolved, absRoot+string(filepath.Separator)) &&
		resolved != absRoot {
		return "", ErrTraversal
	}

	return joined, nil
}
```

**Step 2: Run tests**

Run: `go test ./internal/safepath/ -v`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/safepath/safepath.go
git commit -m "feat: implement SafePath with traversal protection"
```

---

## Task 5: Config — write failing tests

**Files:**
- Create: `internal/config/config.go` (types only)
- Create: `internal/config/config_test.go`

**Step 1: Write test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

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
	t.Setenv("BIBLIUM_DATA_DIR", "/tmp/biblium-test")
	t.Setenv("BIBLIUM_SEARCH_BACKEND", "ollama")
	t.Setenv("BIBLIUM_LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDir != "/tmp/biblium-test" {
		t.Errorf("DataDir got %q, want %q", cfg.DataDir, "/tmp/biblium-test")
	}
	if cfg.SearchBackend != "ollama" {
		t.Errorf("SearchBackend got %q, want %q", cfg.SearchBackend, "ollama")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel got %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestLoadRejectsInvalidBackend(t *testing.T) {
	t.Setenv("BIBLIUM_SEARCH_BACKEND", "invalid")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid backend, got nil")
	}
}

func TestLoadOllamaDefaults(t *testing.T) {
	t.Setenv("BIBLIUM_SEARCH_BACKEND", "ollama")
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
```

**Step 2: Write minimal types**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package config

// Config holds all runtime configuration.
type Config struct {
	DataDir       string
	SearchBackend string
	LogLevel      string
	OllamaURL     string
	OllamaModel   string
}

// Load reads configuration from environment variables with defaults.
func Load() (*Config, error) {
	return nil, nil // TDD: stub
}
```

**Step 3: Run tests to verify they FAIL**

Run: `go test ./internal/config/ -v`
Expected: FAIL (nil config)

**Step 4: Commit**

```bash
git add internal/config/
git commit -m "test: add config tests (RED)"
```

---

## Task 6: Config — implement to make tests green

**Files:**
- Modify: `internal/config/config.go`

**Step 1: Implement Load**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

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
		DataDir:       envOr("BIBLIUM_DATA_DIR", defaultDataDir()),
		SearchBackend: envOr("BIBLIUM_SEARCH_BACKEND", "bm25"),
		LogLevel:      envOr("BIBLIUM_LOG_LEVEL", "info"),
		OllamaURL:     envOr("BIBLIUM_OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:   envOr("BIBLIUM_OLLAMA_MODEL", "all-minilm:l6-v2"),
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
		return "biblium_data"
	}
	return filepath.Join(home, "biblium_data")
}
```

**Step 2: Run tests**

Run: `go test ./internal/config/ -v`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "feat: implement config loader from env vars"
```

---

## Task 7: Corpus types and store — write failing tests

**Files:**
- Create: `internal/corpus/store.go` (types/interfaces only)
- Create: `internal/corpus/store_test.go`

**Step 1: Write test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmpty(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	names, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}
}

func TestCreateAndList(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)

	err := s.Create("infra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[0] != "infra" {
		t.Errorf("expected [infra], got %v", names)
	}
}

func TestCreateDirStructure(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	_ = s.Create("podman")

	docsDir := filepath.Join(root, "corpora", "podman", "docs")
	info, err := os.Stat(docsDir)
	if err != nil {
		t.Fatalf("docs dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("docs should be a directory")
	}
}

func TestCreateRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	err := s.Create("../evil")
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
}

func TestListDocumentsEmpty(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	_ = s.Create("test")

	docs, err := s.ListDocs("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("expected empty docs, got %v", docs)
	}
}

func TestAddAndListDocuments(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	_ = s.Create("test")

	err := s.AddDoc("test", "hello.md", []byte("# Hello\n\nWorld"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	docs, err := s.ListDocs("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 || docs[0] != "hello.md" {
		t.Errorf("expected [hello.md], got %v", docs)
	}
}

func TestReadDocument(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	_ = s.Create("test")
	_ = s.AddDoc("test", "hello.md", []byte("# Hello"))

	content, err := s.ReadDoc("test", "hello.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(content) != "# Hello" {
		t.Errorf("got %q, want %q", string(content), "# Hello")
	}
}

func TestAddDocRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	s := NewFileStore(root)
	_ = s.Create("test")

	err := s.AddDoc("test", "../../evil.md", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for doc path traversal, got nil")
	}
}
```

**Step 2: Write minimal types/interfaces**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package corpus

// FileStore manages corpora on the filesystem.
type FileStore struct {
	root string
}

// NewFileStore creates a store rooted at the given directory.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

func (s *FileStore) List() ([]string, error)                         { return nil, nil }
func (s *FileStore) Create(name string) error                        { return nil }
func (s *FileStore) ListDocs(corpus string) ([]string, error)        { return nil, nil }
func (s *FileStore) AddDoc(corpus, name string, data []byte) error   { return nil }
func (s *FileStore) ReadDoc(corpus, name string) ([]byte, error)     { return nil, nil }
```

**Step 3: Run tests to verify they FAIL**

Run: `go test ./internal/corpus/ -v`
Expected: Multiple FAIL

**Step 4: Commit**

```bash
git add internal/corpus/
git commit -m "test: add corpus store tests (RED)"
```

---

## Task 8: Corpus store — implement to make tests green

**Files:**
- Modify: `internal/corpus/store.go`

**Step 1: Implement all methods**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package corpus

import (
	"os"
	"path/filepath"

	"github.com/mipsou/mcp-biblium/internal/safepath"
)

// FileStore manages corpora on the filesystem.
type FileStore struct {
	root string // root/corpora/<name>/docs/
}

// NewFileStore creates a store rooted at the given directory.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

func (s *FileStore) corporaDir() string {
	return filepath.Join(s.root, "corpora")
}

// List returns the names of all corpora.
func (s *FileStore) List() ([]string, error) {
	dir := s.corporaDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && e.Name()[0] != '.' {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// Create initializes a new corpus directory with a docs/ subdirectory.
func (s *FileStore) Create(name string) error {
	_, err := safepath.Resolve(s.corporaDir(), name)
	if err != nil {
		return err
	}
	docsDir := filepath.Join(s.corporaDir(), name, "docs")
	return os.MkdirAll(docsDir, 0o755)
}

// ListDocs returns the filenames of all documents in a corpus.
func (s *FileStore) ListDocs(corpus string) ([]string, error) {
	_, err := safepath.Resolve(s.corporaDir(), corpus)
	if err != nil {
		return nil, err
	}
	docsDir := filepath.Join(s.corporaDir(), corpus, "docs")
	entries, err := os.ReadDir(docsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// AddDoc writes a document file to a corpus.
func (s *FileStore) AddDoc(corpus, name string, data []byte) error {
	corpusPath := filepath.Join(s.corporaDir(), corpus, "docs")
	resolved, err := safepath.Resolve(corpusPath, name)
	if err != nil {
		return err
	}
	return os.WriteFile(resolved, data, 0o644)
}

// ReadDoc reads a document from a corpus.
func (s *FileStore) ReadDoc(corpus, name string) ([]byte, error) {
	corpusPath := filepath.Join(s.corporaDir(), corpus, "docs")
	resolved, err := safepath.Resolve(corpusPath, name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolved)
}
```

**Step 2: Run tests**

Run: `go test ./internal/corpus/ -v`
Expected: ALL PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/corpus/store.go
git commit -m "feat: implement file-based corpus store"
```

---

## Task 9: BM25 search — write failing tests

**Files:**
- Create: `internal/search/searcher.go` (interface only)
- Create: `internal/search/bm25.go` (types only)
- Create: `internal/search/bm25_test.go`

**Step 1: Write Searcher interface**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package search

// Result represents a single search result.
type Result struct {
	Corpus  string  // corpus name
	DocName string  // document filename
	Score   float64 // relevance score (higher = better)
	Snippet string  // matching text excerpt
}

// Searcher is the interface for all search backends.
type Searcher interface {
	// Index processes a document and adds it to the search index.
	Index(corpus, docName, content string) error

	// Search returns ranked results for a query.
	// maxResults limits the number of results returned.
	Search(query string, maxResults int) ([]Result, error)

	// Remove removes a document from the index.
	Remove(corpus, docName string) error
}
```

**Step 2: Write BM25 test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

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
	// podman.md should rank higher (more term matches)
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
```

**Step 3: Write BM25 stub**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package search

// BM25 implements the Searcher interface using the BM25 ranking algorithm.
// Pure Go, zero external dependencies.
type BM25 struct{}

// NewBM25 creates a new BM25 search engine.
func NewBM25() *BM25 { return &BM25{} }

func (b *BM25) Index(corpus, docName, content string) error          { return nil }
func (b *BM25) Search(query string, maxResults int) ([]Result, error) { return nil, nil }
func (b *BM25) Remove(corpus, docName string) error                  { return nil }
```

**Step 4: Run tests to verify they FAIL**

Run: `go test ./internal/search/ -v`
Expected: Multiple FAIL

**Step 5: Commit**

```bash
git add internal/search/
git commit -m "test: add BM25 search tests (RED)"
```

---

## Task 10: BM25 search — implement to make tests green

**Files:**
- Modify: `internal/search/bm25.go`

**Step 1: Implement BM25**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package search

import (
	"math"
	"sort"
	"strings"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

type docEntry struct {
	corpus  string
	docName string
	terms   map[string]int // term -> frequency
	length  int            // total tokens
	content string         // raw content for snippet
}

// BM25 implements the Searcher interface using the BM25 ranking algorithm.
// Pure Go, zero external dependencies.
type BM25 struct {
	docs   []docEntry
	avgLen float64
}

// NewBM25 creates a new BM25 search engine.
func NewBM25() *BM25 {
	return &BM25{}
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	for _, word := range strings.Fields(text) {
		// Strip basic punctuation
		word = strings.Trim(word, ".,;:!?\"'()[]{}#*")
		if word != "" {
			tokens = append(tokens, word)
		}
	}
	return tokens
}

func (b *BM25) Index(corpus, docName, content string) error {
	tokens := tokenize(content)
	tf := make(map[string]int)
	for _, t := range tokens {
		tf[t]++
	}

	b.docs = append(b.docs, docEntry{
		corpus:  corpus,
		docName: docName,
		terms:   tf,
		length:  len(tokens),
		content: content,
	})
	b.recalcAvgLen()
	return nil
}

func (b *BM25) Remove(corpus, docName string) error {
	for i, d := range b.docs {
		if d.corpus == corpus && d.docName == docName {
			b.docs = append(b.docs[:i], b.docs[i+1:]...)
			b.recalcAvgLen()
			return nil
		}
	}
	return nil
}

func (b *BM25) recalcAvgLen() {
	if len(b.docs) == 0 {
		b.avgLen = 0
		return
	}
	total := 0
	for _, d := range b.docs {
		total += d.length
	}
	b.avgLen = float64(total) / float64(len(b.docs))
}

func (b *BM25) Search(query string, maxResults int) ([]Result, error) {
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil, nil
	}

	n := float64(len(b.docs))
	if n == 0 {
		return nil, nil
	}

	// Count document frequency per query term
	df := make(map[string]int)
	for _, qt := range queryTerms {
		for _, d := range b.docs {
			if d.terms[qt] > 0 {
				df[qt]++
			}
		}
	}

	type scored struct {
		idx   int
		score float64
	}
	var scores []scored

	for i, d := range b.docs {
		score := 0.0
		for _, qt := range queryTerms {
			dFreq := float64(df[qt])
			if dFreq == 0 {
				continue
			}
			// IDF component
			idf := math.Log((n-dFreq+0.5) / (dFreq + 0.5))
			if idf < 0 {
				idf = 0
			}

			// TF component
			tf := float64(d.terms[qt])
			dl := float64(d.length)
			tfNorm := (tf * (bm25K1 + 1)) /
				(tf + bm25K1*(1-bm25B+bm25B*dl/b.avgLen))

			score += idf * tfNorm
		}
		if score > 0 {
			scores = append(scores, scored{i, score})
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(a, z int) bool {
		return scores[a].score > scores[z].score
	})

	if len(scores) > maxResults {
		scores = scores[:maxResults]
	}

	results := make([]Result, len(scores))
	for i, s := range scores {
		d := b.docs[s.idx]
		snippet := d.content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results[i] = Result{
			Corpus:  d.corpus,
			DocName: d.docName,
			Score:   s.score,
			Snippet: snippet,
		}
	}
	return results, nil
}
```

**Step 2: Run tests**

Run: `go test ./internal/search/ -v`
Expected: ALL PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/search/bm25.go
git commit -m "feat: implement BM25 search engine (pure Go)"
```

---

## Task 11: MCP server — write failing tests

**Files:**
- Create: `internal/mcp/server.go` (types only)
- Create: `internal/mcp/server_test.go`

**Step 1: Add mcp-go dependency**

```bash
go get github.com/mark3labs/mcp-go@latest
```

**Step 2: Write test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package mcp

import (
	"testing"
)

func TestNewServerNotNil(t *testing.T) {
	root := t.TempDir()
	s, err := NewBibliumServer(root, "bm25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestServerHasTools(t *testing.T) {
	root := t.TempDir()
	s, err := NewBibliumServer(root, "bm25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := s.ToolNames()
	expected := []string{"list_corpora", "search", "suggest", "pending", "approve", "reject", "ingest"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d tools, got %d: %v", len(expected), len(names), names)
	}
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, e := range expected {
		if !nameSet[e] {
			t.Errorf("missing tool: %s", e)
		}
	}
}
```

**Step 3: Write stub**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package mcp

// BibliumServer wraps the MCP server with corpus + search capabilities.
type BibliumServer struct{}

// NewBibliumServer creates and configures the MCP server.
func NewBibliumServer(dataDir, searchBackend string) (*BibliumServer, error) {
	return nil, nil
}

// ToolNames returns the names of all registered tools.
func (s *BibliumServer) ToolNames() []string { return nil }
```

**Step 4: Run tests to verify they FAIL**

Run: `go test ./internal/mcp/ -v`
Expected: FAIL

**Step 5: Commit**

```bash
git add internal/mcp/
git commit -m "test: add MCP server tests (RED)"
```

---

## Task 12: MCP server — implement with tool registration

**Files:**
- Modify: `internal/mcp/server.go`

**Step 1: Implement BibliumServer with mcp-go**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mipsou/mcp-biblium/internal/corpus"
	"github.com/mipsou/mcp-biblium/internal/search"
)

// BibliumServer wraps the MCP server with corpus + search capabilities.
type BibliumServer struct {
	mcpServer *server.MCPServer
	store     *corpus.FileStore
	searcher  search.Searcher
	tools     []string
}

// NewBibliumServer creates and configures the MCP server.
func NewBibliumServer(dataDir, searchBackend string) (*BibliumServer, error) {
	store := corpus.NewFileStore(dataDir)

	var searcher search.Searcher
	switch searchBackend {
	case "bm25":
		searcher = search.NewBM25()
	default:
		searcher = search.NewBM25()
	}

	mcpSrv := server.NewMCPServer(
		"biblium",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	ls := &BibliumServer{
		mcpServer: mcpSrv,
		store:     store,
		searcher:  searcher,
	}

	ls.registerTools()
	return ls, nil
}

func (s *BibliumServer) registerTools() {
	toolDefs := []struct {
		name string
		desc string
		schema mcp.ToolInputSchema
		handler server.ToolHandlerFunc
	}{
		{
			name: "list_corpora",
			desc: "List all available corpora.",
			schema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]map[string]interface{}{},
			},
			handler: s.handleListCorpora,
		},
		{
			name: "search",
			desc: "Search across corpora using a text query.",
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"query":      {"type": "string", "description": "Search query text."},
					"corpus":     {"type": "string", "description": "Optional corpus name to search within."},
					"max_results": {"type": "number", "description": "Maximum results to return (default 10)."},
				},
				Required: []string{"query"},
			},
			handler: s.handleSearch,
		},
		{
			name: "suggest",
			desc: "Suggest a URL for indexing into a corpus (pending approval).",
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"url":    {"type": "string", "description": "URL to suggest."},
					"corpus": {"type": "string", "description": "Target corpus name."},
					"reason": {"type": "string", "description": "Why this URL is relevant."},
				},
				Required: []string{"url", "corpus", "reason"},
			},
			handler: s.handleSuggest,
		},
		{
			name: "pending",
			desc: "List all URLs pending approval.",
			schema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]map[string]interface{}{},
			},
			handler: s.handlePending,
		},
		{
			name: "approve",
			desc: "Approve a pending URL: fetch, convert, and index.",
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"id": {"type": "string", "description": "Pending URL id."},
				},
				Required: []string{"id"},
			},
			handler: s.handleApprove,
		},
		{
			name: "reject",
			desc: "Reject and remove a pending URL.",
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"id": {"type": "string", "description": "Pending URL id."},
				},
				Required: []string{"id"},
			},
			handler: s.handleReject,
		},
		{
			name: "ingest",
			desc: "Directly fetch a URL, convert to markdown, and index.",
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]map[string]interface{}{
					"url":    {"type": "string", "description": "URL to fetch."},
					"corpus": {"type": "string", "description": "Target corpus name."},
				},
				Required: []string{"url", "corpus"},
			},
			handler: s.handleIngest,
		},
	}

	for _, td := range toolDefs {
		s.tools = append(s.tools, td.name)
		s.mcpServer.AddTool(mcp.Tool{
			Name:        td.name,
			Description: td.desc,
			InputSchema: td.schema,
		}, td.handler)
	}
}

// ToolNames returns the names of all registered tools.
func (s *BibliumServer) ToolNames() []string {
	return s.tools
}

// Serve starts the MCP server on stdio.
func (s *BibliumServer) Serve() error {
	return server.ServeStdio(s.mcpServer)
}

// Stub handlers — will be implemented in Task 13+
func (s *BibliumServer) handleListCorpora(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("[]"), nil
}
func (s *BibliumServer) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("[]"), nil
}
func (s *BibliumServer) handleSuggest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}
func (s *BibliumServer) handlePending(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("[]"), nil
}
func (s *BibliumServer) handleApprove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}
func (s *BibliumServer) handleReject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}
func (s *BibliumServer) handleIngest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}
```

**Step 2: Run tests**

Run: `go test ./internal/mcp/ -v`
Expected: ALL PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/mcp/ go.mod go.sum
git commit -m "feat: MCP server skeleton with 7 tool registrations"
```

---

## Task 13: Wire main.go to MCP server

**Files:**
- Modify: `cmd/biblium/main.go`

**Step 1: Update main.go**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mipsou/mcp-biblium/internal/config"
	bibmcp "github.com/mipsou/mcp-biblium/internal/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "biblium: config error: %v\n", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	slog.Info("biblium starting", "data_dir", cfg.DataDir, "search_backend", cfg.SearchBackend)

	srv, err := bibmcp.NewBibliumServer(cfg.DataDir, cfg.SearchBackend)
	if err != nil {
		slog.Error("server init failed", "error", err)
		os.Exit(1)
	}

	if err := srv.Serve(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
```

**Step 2: Build and verify**

Run: `go build -o biblium ./cmd/biblium/`
Expected: binary compiles, ~10-15MB

Run: `ls -lh biblium`
Expected: single binary file

**Step 3: Commit**

```bash
git add cmd/biblium/main.go
git commit -m "feat: wire main.go to MCP server"
```

---

## Task 14: Implement list_corpora and search tool handlers — tests first

**Files:**
- Create: `internal/mcp/handlers_test.go`
- Modify: `internal/mcp/server.go`

**Step 1: Write handler integration tests**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mipsou/mcp-biblium/internal/corpus"
)

func setupTestServer(t *testing.T) *BibliumServer {
	t.Helper()
	root := t.TempDir()
	s, err := NewBibliumServer(root, "bm25")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	// Create test corpus with a doc
	_ = s.store.Create("infra")
	_ = s.store.AddDoc("infra", "caddy.md", []byte("Caddy is a web server with automatic HTTPS"))
	return s
}

func callTool(t *testing.T, s *BibliumServer, name string, args map[string]interface{}) string {
	t.Helper()
	argsJSON, _ := json.Marshal(args)
	var rawArgs map[string]interface{}
	_ = json.Unmarshal(argsJSON, &rawArgs)

	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = rawArgs

	var handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	switch name {
	case "list_corpora":
		handler = s.handleListCorpora
	case "search":
		handler = s.handleSearch
	default:
		t.Fatalf("unknown tool: %s", name)
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if len(result.Content) == 0 {
		return ""
	}
	tc := result.Content[0].(mcp.TextContent)
	return tc.Text
}

func TestHandleListCorporaReturnsCorpora(t *testing.T) {
	s := setupTestServer(t)
	text := callTool(t, s, "list_corpora", nil)
	if !strings.Contains(text, "infra") {
		t.Errorf("expected 'infra' in result, got: %s", text)
	}
}

func TestHandleSearchFindsDocument(t *testing.T) {
	s := setupTestServer(t)
	// First, index the docs
	_ = s.indexCorpus("infra")

	text := callTool(t, s, "search", map[string]interface{}{
		"query": "caddy HTTPS",
	})
	if !strings.Contains(text, "caddy") {
		t.Errorf("expected 'caddy' in search result, got: %s", text)
	}
}

func TestHandleSearchEmptyQuery(t *testing.T) {
	s := setupTestServer(t)
	text := callTool(t, s, "search", map[string]interface{}{
		"query": "",
	})
	if strings.Contains(text, "caddy") {
		t.Errorf("expected no results for empty query, got: %s", text)
	}
}
```

**Step 2: Run tests to verify they FAIL**

Run: `go test ./internal/mcp/ -v -run TestHandle`
Expected: FAIL (stub returns "[]", indexCorpus doesn't exist)

**Step 3: Commit**

```bash
git add internal/mcp/handlers_test.go
git commit -m "test: add tool handler integration tests (RED)"
```

---

## Task 15: Implement list_corpora and search handlers

**Files:**
- Modify: `internal/mcp/server.go`

**Step 1: Add indexCorpus helper and implement handlers**

Add `indexCorpus` method and update `handleListCorpora` and `handleSearch`:

```go
// indexCorpus reads all docs from a corpus and indexes them for search.
func (s *BibliumServer) indexCorpus(name string) error {
	docs, err := s.store.ListDocs(name)
	if err != nil {
		return err
	}
	for _, docName := range docs {
		content, err := s.store.ReadDoc(name, docName)
		if err != nil {
			continue
		}
		_ = s.searcher.Index(name, docName, string(content))
	}
	return nil
}

// indexAllCorpora indexes all corpora.
func (s *BibliumServer) indexAllCorpora() error {
	names, err := s.store.List()
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := s.indexCorpus(name); err != nil {
			return err
		}
	}
	return nil
}

func (s *BibliumServer) handleListCorpora(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	names, err := s.store.List()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if names == nil {
		names = []string{}
	}
	data, _ := json.Marshal(names)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *BibliumServer) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := req.Params.Arguments["query"].(string)
	if query == "" {
		return mcp.NewToolResultText("[]"), nil
	}

	maxResults := 10
	if mr, ok := req.Params.Arguments["max_results"].(float64); ok && mr > 0 {
		maxResults = int(mr)
	}

	results, err := s.searcher.Search(query, maxResults)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.Marshal(results)
	return mcp.NewToolResultText(string(data)), nil
}
```

Add `"encoding/json"` to imports.

**Step 2: Run tests**

Run: `go test ./internal/mcp/ -v`
Expected: ALL PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/mcp/server.go
git commit -m "feat: implement list_corpora and search tool handlers"
```

---

## Task 16: URL ingestion — write failing tests

**Files:**
- Create: `internal/ingest/ingest.go` (types only)
- Create: `internal/ingest/ingest_test.go`

**Step 1: Write test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAndConvert(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Hello</h1><p>World</p></body></html>"))
	}))
	defer srv.Close()

	md, err := FetchAndConvert(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md == "" {
		t.Fatal("expected non-empty markdown")
	}
	// Should contain the heading
	if !containsAny(md, "Hello") {
		t.Errorf("expected 'Hello' in markdown, got: %s", md)
	}
}

func TestFetchRejectsPrivateIP(t *testing.T) {
	_, err := FetchAndConvert("http://127.0.0.1:1234/secret")
	if err == nil {
		t.Fatal("expected error for private IP, got nil")
	}
}

func TestFetchRejectsNonHTTP(t *testing.T) {
	_, err := FetchAndConvert("ftp://example.com/file")
	if err == nil {
		t.Fatal("expected error for non-HTTP scheme, got nil")
	}
}

func TestFetchRejectsEmptyURL(t *testing.T) {
	_, err := FetchAndConvert("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func containsAny(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 &&
		(haystack == needle || len(haystack) > len(needle) &&
			(haystack[:len(needle)] == needle ||
				containsAny(haystack[1:], needle)))
}
```

Note: replace the `containsAny` helper with `strings.Contains` — simpler:

```go
import "strings"
// In test: strings.Contains(md, "Hello")
```

**Step 2: Write stub**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

// FetchAndConvert downloads a URL and returns markdown content.
func FetchAndConvert(rawURL string) (string, error) {
	return "", nil
}
```

**Step 3: Run tests to verify they FAIL**

Run: `go test ./internal/ingest/ -v`
Expected: FAIL

**Step 4: Commit**

```bash
git add internal/ingest/
git commit -m "test: add URL ingestion tests (RED)"
```

---

## Task 17: URL ingestion — implement

**Files:**
- Modify: `internal/ingest/ingest.go`

**Step 1: Add html-to-markdown dependency**

```bash
go get github.com/JohannesKaufmann/html-to-markdown/v2@latest
```

**Step 2: Implement FetchAndConvert**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	htmltomd "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	ErrEmptyURL    = errors.New("ingest: empty URL")
	ErrBadScheme   = errors.New("ingest: only http/https allowed")
	ErrPrivateIP   = errors.New("ingest: private/loopback IP not allowed")
	ErrTooLarge    = errors.New("ingest: response exceeds 10MB limit")
)

const maxResponseSize = 10 * 1024 * 1024 // 10MB

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// FetchAndConvert downloads a URL and returns markdown content.
func FetchAndConvert(rawURL string) (string, error) {
	if rawURL == "" {
		return "", ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("ingest: invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrBadScheme
	}

	// Check for private IPs
	host := parsed.Hostname()
	if isPrivateHost(host) {
		return "", ErrPrivateIP
	}

	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("ingest: fetch failed: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("ingest: read failed: %w", err)
	}
	if len(body) > maxResponseSize {
		return "", ErrTooLarge
	}

	md, err := htmltomd.ConvertString(string(body))
	if err != nil {
		return "", fmt.Errorf("ingest: conversion failed: %w", err)
	}

	return strings.TrimSpace(md), nil
}

func isPrivateHost(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		// Try resolving hostname
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return false
		}
		ip = ips[0]
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}
```

**Step 2: Run tests**

Run: `go test ./internal/ingest/ -v`
Expected: ALL PASS (note: `TestFetchRejectsPrivateIP` tests 127.0.0.1)

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/ingest/ go.mod go.sum
git commit -m "feat: implement URL fetch + HTML to markdown conversion"
```

---

## Task 18: Pending URL approval workflow — tests first

**Files:**
- Create: `internal/ingest/pending.go` (types)
- Create: `internal/ingest/pending_test.go`

**Step 1: Add uuid dependency**

```bash
go get github.com/google/uuid
```

**Step 2: Write test file**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

import (
	"path/filepath"
	"testing"
)

func TestSuggestAndList(t *testing.T) {
	dir := t.TempDir()
	p := NewPendingStore(filepath.Join(dir, "pending.json"))

	entry, err := p.Suggest("https://example.com", "infra", "useful doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if entry.URL != "https://example.com" {
		t.Errorf("URL mismatch: %s", entry.URL)
	}

	list, err := p.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(list))
	}
}

func TestApproveRemovesFromPending(t *testing.T) {
	dir := t.TempDir()
	p := NewPendingStore(filepath.Join(dir, "pending.json"))

	entry, _ := p.Suggest("https://example.com", "infra", "test")
	got, err := p.Approve(entry.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.URL != "https://example.com" {
		t.Errorf("URL mismatch: %s", got.URL)
	}

	list, _ := p.List()
	if len(list) != 0 {
		t.Errorf("expected empty list after approve, got %d", len(list))
	}
}

func TestRejectRemovesFromPending(t *testing.T) {
	dir := t.TempDir()
	p := NewPendingStore(filepath.Join(dir, "pending.json"))

	entry, _ := p.Suggest("https://example.com", "infra", "test")
	err := p.Reject(entry.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list, _ := p.List()
	if len(list) != 0 {
		t.Errorf("expected empty list after reject, got %d", len(list))
	}
}

func TestApproveUnknownID(t *testing.T) {
	dir := t.TempDir()
	p := NewPendingStore(filepath.Join(dir, "pending.json"))

	_, err := p.Approve("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown ID, got nil")
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pending.json")

	p1 := NewPendingStore(file)
	_, _ = p1.Suggest("https://example.com", "infra", "test")

	// Create new instance pointing at same file
	p2 := NewPendingStore(file)
	list, err := p2.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected persisted entry, got %d", len(list))
	}
}
```

**Step 3: Write stub**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

// PendingEntry represents a URL awaiting approval.
type PendingEntry struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Corpus  string `json:"corpus"`
	Reason  string `json:"reason"`
	AddedAt string `json:"added_at"`
}

// PendingStore manages the pending URL queue.
type PendingStore struct {
	path string
}

func NewPendingStore(path string) *PendingStore         { return &PendingStore{path: path} }
func (p *PendingStore) Suggest(url, corpus, reason string) (*PendingEntry, error) { return nil, nil }
func (p *PendingStore) List() ([]PendingEntry, error)    { return nil, nil }
func (p *PendingStore) Approve(id string) (*PendingEntry, error) { return nil, nil }
func (p *PendingStore) Reject(id string) error           { return nil }
```

**Step 4: Run tests to verify they FAIL**

Run: `go test ./internal/ingest/ -v -run TestSuggest`
Expected: FAIL

**Step 5: Commit**

```bash
git add internal/ingest/pending.go internal/ingest/pending_test.go go.mod go.sum
git commit -m "test: add pending URL workflow tests (RED)"
```

---

## Task 19: Pending URL workflow — implement

**Files:**
- Modify: `internal/ingest/pending.go`

**Step 1: Implement PendingStore**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package ingest

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("ingest: pending entry not found")

// PendingEntry represents a URL awaiting approval.
type PendingEntry struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Corpus  string `json:"corpus"`
	Reason  string `json:"reason"`
	AddedAt string `json:"added_at"`
}

// PendingStore manages the pending URL queue (JSON file).
type PendingStore struct {
	path string
}

func NewPendingStore(path string) *PendingStore {
	return &PendingStore{path: path}
}

func (p *PendingStore) load() ([]PendingEntry, error) {
	data, err := os.ReadFile(p.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []PendingEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (p *PendingStore) save(entries []PendingEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.path, data, 0o644)
}

func (p *PendingStore) Suggest(rawURL, corpus, reason string) (*PendingEntry, error) {
	entries, err := p.load()
	if err != nil {
		return nil, err
	}

	entry := PendingEntry{
		ID:      uuid.New().String(),
		URL:     rawURL,
		Corpus:  corpus,
		Reason:  reason,
		AddedAt: time.Now().UTC().Format(time.RFC3339),
	}
	entries = append(entries, entry)
	if err := p.save(entries); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (p *PendingStore) List() ([]PendingEntry, error) {
	return p.load()
}

func (p *PendingStore) Approve(id string) (*PendingEntry, error) {
	entries, err := p.load()
	if err != nil {
		return nil, err
	}
	for i, e := range entries {
		if e.ID == id {
			entries = append(entries[:i], entries[i+1:]...)
			if err := p.save(entries); err != nil {
				return nil, err
			}
			return &e, nil
		}
	}
	return nil, ErrNotFound
}

func (p *PendingStore) Reject(id string) error {
	entries, err := p.load()
	if err != nil {
		return err
	}
	for i, e := range entries {
		if e.ID == id {
			entries = append(entries[:i], entries[i+1:]...)
			return p.save(entries)
		}
	}
	return ErrNotFound
}
```

**Step 2: Run tests**

Run: `go test ./internal/ingest/ -v`
Expected: ALL PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/ingest/pending.go
git commit -m "feat: implement pending URL approval workflow"
```

---

## Task 20: Wire remaining tool handlers + final integration

**Files:**
- Modify: `internal/mcp/server.go` (wire suggest/pending/approve/reject/ingest handlers)
- Create: `internal/mcp/integration_test.go`

**Step 1: Write integration test**

```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>

package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestFullWorkflow(t *testing.T) {
	root := t.TempDir()
	s, err := NewBibliumServer(root, "bm25")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// 1. list_corpora — should be empty
	res := callToolDirect(t, s, "list_corpora", nil)
	if res != "[]" {
		t.Errorf("expected empty list, got: %s", res)
	}

	// 2. Create a corpus and add doc manually for search test
	_ = s.store.Create("infra")
	_ = s.store.AddDoc("infra", "caddy.md", []byte("Caddy is a web server with automatic HTTPS"))
	_ = s.indexCorpus("infra")

	// 3. list_corpora — should have infra
	res = callToolDirect(t, s, "list_corpora", nil)
	if !strings.Contains(res, "infra") {
		t.Errorf("expected 'infra', got: %s", res)
	}

	// 4. search
	res = callToolDirect(t, s, "search", map[string]interface{}{
		"query": "web server HTTPS",
	})
	if !strings.Contains(strings.ToLower(res), "caddy") {
		t.Errorf("expected caddy in results, got: %s", res)
	}
}

func callToolDirect(t *testing.T, s *BibliumServer, name string, args map[string]interface{}) string {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args

	handlers := map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){
		"list_corpora": s.handleListCorpora,
		"search":       s.handleSearch,
		"suggest":      s.handleSuggest,
		"pending":      s.handlePending,
		"approve":      s.handleApprove,
		"reject":       s.handleReject,
		"ingest":       s.handleIngest,
	}

	handler, ok := handlers[name]
	if !ok {
		t.Fatalf("unknown tool: %s", name)
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if len(result.Content) == 0 {
		return ""
	}
	data, _ := json.Marshal(result.Content[0])
	var tc struct{ Text string }
	_ = json.Unmarshal(data, &tc)
	return tc.Text
}
```

**Step 2: Update server.go to wire pending/suggest/approve/reject/ingest**

Add `PendingStore` to `BibliumServer` and implement the remaining handlers using
`internal/ingest` package. Each handler:
- Extracts params from `req.Params.Arguments`
- Calls the appropriate store/ingest method
- Returns JSON result via `mcp.NewToolResultText`

**Step 3: Run all tests**

Run: `go test ./... -v -count=1`
Expected: ALL PASS

**Step 4: Build final binary**

Run: `go build -o biblium ./cmd/biblium/ && ls -lh biblium`
Expected: single binary, <20MB

**Step 5: Commit**

```bash
git add internal/mcp/
git commit -m "feat: wire all 7 MCP tool handlers — MVP complete"
```

---

## Summary

| Task | Description | TDD Phase | Key Files |
|------|-------------|-----------|-----------|
| 1 | Branch + Go module + license | Setup | go.mod, LICENSE |
| 2 | main.go entry point | Setup | cmd/biblium/main.go |
| 3 | SafePath tests | RED | internal/safepath/*_test.go |
| 4 | SafePath implementation | GREEN | internal/safepath/safepath.go |
| 5 | Config tests | RED | internal/config/*_test.go |
| 6 | Config implementation | GREEN | internal/config/config.go |
| 7 | Corpus store tests | RED | internal/corpus/*_test.go |
| 8 | Corpus store implementation | GREEN | internal/corpus/store.go |
| 9 | BM25 search tests | RED | internal/search/*_test.go |
| 10 | BM25 implementation | GREEN | internal/search/bm25.go |
| 11 | MCP server tests | RED | internal/mcp/*_test.go |
| 12 | MCP server skeleton | GREEN | internal/mcp/server.go |
| 13 | Wire main.go | Setup | cmd/biblium/main.go |
| 14 | Handler tests | RED | internal/mcp/handlers_test.go |
| 15 | Handler implementation | GREEN | internal/mcp/server.go |
| 16 | Ingestion tests | RED | internal/ingest/*_test.go |
| 17 | Ingestion implementation | GREEN | internal/ingest/ingest.go |
| 18 | Pending workflow tests | RED | internal/ingest/pending_test.go |
| 19 | Pending workflow implementation | GREEN | internal/ingest/pending.go |
| 20 | Final wiring + integration | GREEN | internal/mcp/server.go |

**Total estimated Go files: ~15**
**Total estimated lines: ~2500**
**External dependencies: 3** (mcp-go, html-to-markdown, uuid)
**Binary: single static binary, <20MB**
