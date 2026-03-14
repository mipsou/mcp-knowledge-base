/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

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
