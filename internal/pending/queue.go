/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package pending

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
type Queue struct{}

// NewQueue creates a new pending queue.
func NewQueue() *Queue { return &Queue{} }

// Add adds a URL to the pending queue.
func (q *Queue) Add(corpus, rawURL string) (*Entry, error) { return nil, nil }

// List returns all pending entries.
func (q *Queue) List() []*Entry { return nil }

// Approve marks an entry as approved by ID.
func (q *Queue) Approve(id string) (*Entry, error) { return nil, nil }

// Reject marks an entry as rejected by ID.
func (q *Queue) Reject(id string) (*Entry, error) { return nil, nil }
