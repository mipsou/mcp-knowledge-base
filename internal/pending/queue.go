/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package pending

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Status represents the state of a pending URL.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Entry is a URL waiting for approval.
type Entry struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Corpus string `json:"corpus"`
	Status Status `json:"status"`
}

// Queue manages pending URL suggestions.
type Queue struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// NewQueue creates a new pending queue.
func NewQueue() *Queue {
	return &Queue{
		entries: make(map[string]*Entry),
	}
}

// Add adds a URL to the pending queue.
func (q *Queue) Add(corpus, rawURL string) (*Entry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entry := &Entry{
		ID:     uuid.New().String(),
		URL:    rawURL,
		Corpus: corpus,
		Status: StatusPending,
	}
	q.entries[entry.ID] = entry
	return entry, nil
}

// List returns all pending entries.
func (q *Queue) List() []*Entry {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Entry, 0, len(q.entries))
	for _, e := range q.entries {
		result = append(result, e)
	}
	return result
}

// Approve marks an entry as approved by ID.
func (q *Queue) Approve(id string) (*Entry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entry, ok := q.entries[id]
	if !ok {
		return nil, fmt.Errorf("pending: entry %q not found", id)
	}
	entry.Status = StatusApproved
	return entry, nil
}

// Reject marks an entry as rejected by ID.
func (q *Queue) Reject(id string) (*Entry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entry, ok := q.entries[id]
	if !ok {
		return nil, fmt.Errorf("pending: entry %q not found", id)
	}
	entry.Status = StatusRejected
	return entry, nil
}
