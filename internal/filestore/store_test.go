/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package filestore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmpty(t *testing.T) {
	root := t.TempDir()
	s := New(root)
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
	s := New(root)

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
	s := New(root)
	_ = s.Create("podman")

	docsDir := filepath.Join(root, "podman", "docs")
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
	s := New(root)
	err := s.Create("../evil")
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
}

func TestListDocumentsEmpty(t *testing.T) {
	root := t.TempDir()
	s := New(root)
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
	s := New(root)
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
	s := New(root)
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
	s := New(root)
	_ = s.Create("test")

	err := s.AddDoc("test", "../../evil.md", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for doc path traversal, got nil")
	}
}

func TestWalkEmpty(t *testing.T) {
	root := t.TempDir()
	s := New(root)

	var count int
	err := s.Walk(func(collection, docName, content string) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 docs, got %d", count)
	}
}

func TestWalkVisitsAllDocs(t *testing.T) {
	root := t.TempDir()
	s := New(root)
	_ = s.Create("infra")
	_ = s.Create("golang")
	_ = s.AddDoc("infra", "a.md", []byte("doc A"))
	_ = s.AddDoc("infra", "b.md", []byte("doc B"))
	_ = s.AddDoc("golang", "c.md", []byte("doc C"))

	type doc struct{ collection, name, content string }
	var docs []doc
	err := s.Walk(func(collection, docName, content string) error {
		docs = append(docs, doc{collection, docName, content})
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(docs))
	}
}
