/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package corpus

import (
	"os"
	"path/filepath"

	"github.com/mipsou/lore-mcp/internal/safepath"
)

// FileStore manages corpora on the filesystem.
type FileStore struct {
	root string
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
