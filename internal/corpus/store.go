/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package corpus

// FileStore manages corpora on the filesystem.
type FileStore struct {
	root string
}

// NewFileStore creates a store rooted at the given directory.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

func (s *FileStore) List() ([]string, error)                       { return nil, nil }
func (s *FileStore) Create(name string) error                      { return nil }
func (s *FileStore) ListDocs(corpus string) ([]string, error)      { return nil, nil }
func (s *FileStore) AddDoc(corpus, name string, data []byte) error { return nil }
func (s *FileStore) ReadDoc(corpus, name string) ([]byte, error)   { return nil, nil }
