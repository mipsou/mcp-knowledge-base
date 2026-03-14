/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package filestore

import (
	"os"
	"path/filepath"

	"github.com/mipsou/mcp-biblium/internal/safepath"
)

// Store manages collections on the filesystem.
// Each collection is a directory under root containing a docs/ subdirectory.
type Store struct {
	root string
}

// New creates a store rooted at the given directory.
func New(root string) *Store {
	return &Store{root: root}
}

// List returns the names of all collections.
func (s *Store) List() ([]string, error) {
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

// Create initializes a new collection directory with a docs/ subdirectory.
func (s *Store) Create(name string) error {
	_, err := safepath.Resolve(s.root, name)
	if err != nil {
		return err
	}
	docsDir := filepath.Join(s.root, name, "docs")
	return os.MkdirAll(docsDir, 0o750)
}

// ListDocs returns the filenames of all documents in a collection.
func (s *Store) ListDocs(collection string) ([]string, error) {
	_, err := safepath.Resolve(s.root, collection)
	if err != nil {
		return nil, err
	}
	docsDir := filepath.Join(s.root, collection, "docs")
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

// AddDoc writes a document file to a collection.
func (s *Store) AddDoc(collection, name string, data []byte) error {
	collPath := filepath.Join(s.root, collection, "docs")
	resolved, err := safepath.Resolve(collPath, name)
	if err != nil {
		return err
	}
	return os.WriteFile(resolved, data, 0o600)
}

// ReadDoc reads a document from a collection.
func (s *Store) ReadDoc(collection, name string) ([]byte, error) {
	collPath := filepath.Join(s.root, collection, "docs")
	resolved, err := safepath.Resolve(collPath, name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolved) // #nosec G304 — path validated by safepath.Resolve
}

// Walk iterates over all documents in every collection, calling fn for each.
func (s *Store) Walk(fn func(collection, docName, content string) error) error {
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
