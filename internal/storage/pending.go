/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * Licensed under the EUPL, Version 1.2 or later.
 * You may obtain a copy at:
 * https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
 */

package storage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mipsou/mcp-biblium/internal/pending"
)

// PendingAdd inserts a new pending URL and returns the entry.
func (d *DB) PendingAdd(corpus, rawURL string) (*pending.Entry, error) {
	id := uuid.New().String()
	_, err := d.db.Exec(
		`INSERT INTO pending_urls (id, url, corpus, status) VALUES (?, ?, ?, ?)`,
		id, rawURL, corpus, string(pending.StatusPending),
	)
	if err != nil {
		return nil, err
	}
	return &pending.Entry{
		ID:     id,
		URL:    rawURL,
		Corpus: corpus,
		Status: pending.StatusPending,
	}, nil
}

// PendingList returns all pending entries.
func (d *DB) PendingList() ([]*pending.Entry, error) {
	rows, err := d.db.Query(
		`SELECT id, url, corpus, status FROM pending_urls ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*pending.Entry
	for rows.Next() {
		var e pending.Entry
		var status string
		if err := rows.Scan(&e.ID, &e.URL, &e.Corpus, &status); err != nil {
			return nil, err
		}
		e.Status = pending.Status(status)
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

// PendingApprove marks an entry as approved.
func (d *DB) PendingApprove(id string) (*pending.Entry, error) {
	return d.pendingSetStatus(id, pending.StatusApproved)
}

// PendingReject marks an entry as rejected.
func (d *DB) PendingReject(id string) (*pending.Entry, error) {
	return d.pendingSetStatus(id, pending.StatusRejected)
}

func (d *DB) pendingSetStatus(id string, status pending.Status) (*pending.Entry, error) {
	res, err := d.db.Exec(
		`UPDATE pending_urls SET status = ?, updated_at = datetime('now') WHERE id = ?`,
		string(status), id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("pending: entry %q not found", id)
	}

	var e pending.Entry
	var s string
	err = d.db.QueryRow(
		`SELECT id, url, corpus, status FROM pending_urls WHERE id = ?`, id,
	).Scan(&e.ID, &e.URL, &e.Corpus, &s)
	if err != nil {
		return nil, err
	}
	e.Status = pending.Status(s)
	return &e, nil
}
