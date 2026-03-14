/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

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

	if filepath.IsAbs(userPath) || strings.HasPrefix(userPath, "/") {
		return "", ErrAbsolutePath
	}

	cleaned := filepath.Clean(userPath)

	if cleaned == "." {
		return "", ErrEmptyPath
	}

	if strings.HasPrefix(cleaned, "..") {
		return "", ErrTraversal
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	joined := filepath.Join(absRoot, cleaned)

	// Evaluate symlinks to detect escapes
	resolved, err := filepath.EvalSymlinks(filepath.Dir(joined))
	if err != nil {
		// Parent dir doesn't exist yet — verify cleaned path only
		resolved = joined
	} else {
		resolved = filepath.Join(resolved, filepath.Base(joined))
	}

	if !strings.HasPrefix(resolved, absRoot+string(filepath.Separator)) &&
		resolved != absRoot {
		return "", ErrTraversal
	}

	return joined, nil
}
