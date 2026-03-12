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

	"github.com/mipsou/mcp-biblium/internal/safepath"
)

// FileStore manages corpus entries on the filesystem.
// Each corpus is a directory under root containing a docs/ subdirectory.
type FileStore struct {
	root string
}

// NewFileStore creates a store rooted at the given directory.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

// List returns the names of all corpus entries.
func (s *FileStore) List() ([]string, error) {
	entries, err := os.ReadDir(s.root)
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
	_, err := safepath.Resolve(s.root, name)
	if err != nil {
		return err
	}
	docsDir := filepath.Join(s.root, name, "docs")
	return os.MkdirAll(docsDir, 0o755)
}

// ListDocs returns the filenames of all documents in a corpus.
func (s *FileStore) ListDocs(corpus string) ([]string, error) {
	_, err := safepath.Resolve(s.root, corpus)
	if err != nil {
		return nil, err
	}
	docsDir := filepath.Join(s.root, corpus, "docs")
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
	corpusPath := filepath.Join(s.root, corpus, "docs")
	resolved, err := safepath.Resolve(corpusPath, name)
	if err != nil {
		return err
	}
	return os.WriteFile(resolved, data, 0o644)
}

// ReadDoc reads a document from a corpus.
func (s *FileStore) ReadDoc(corpus, name string) ([]byte, error) {
	corpusPath := filepath.Join(s.root, corpus, "docs")
	resolved, err := safepath.Resolve(corpusPath, name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolved)
}

// Walk iterates over all documents in every corpus, calling fn for each.
func (s *FileStore) Walk(fn func(corpus, docName, content string) error) error {
	names, err := s.List()
	if err != nil {
		return err
	}
	for _, c := range names {
		docs, err := s.ListDocs(c)
		if err != nil {
			return err
		}
		for _, d := range docs {
			data, err := s.ReadDoc(c, d)
			if err != nil {
				return err
			}
			if err := fn(c, d, string(data)); err != nil {
				return err
			}
		}
	}
	return nil
}
