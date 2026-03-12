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
