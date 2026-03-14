/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2 OR BSD-2-Clause
 */

package ingest

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	ErrInvalidURL  = errors.New("ingest: invalid URL")
	ErrBadScheme   = errors.New("ingest: only http and https schemes are allowed")
	ErrSSRF        = errors.New("ingest: request to private/loopback address blocked")
	ErrBadStatus   = errors.New("ingest: non-2xx response")
	ErrBodyTooLong = errors.New("ingest: response body exceeds limit")
)

const maxBodySize = 5 * 1024 * 1024 // 5 MiB

// Fetcher downloads a URL and converts it to markdown.
type Fetcher struct {
	client       *http.Client
	allowPrivate bool // only for testing — bypasses SSRF check
}

// NewFetcher creates a new Fetcher with safe defaults.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newTestFetcher creates a Fetcher that allows private IPs (for httptest).
func newTestFetcher() *Fetcher {
	f := NewFetcher()
	f.allowPrivate = true
	return f
}

// Fetch downloads the URL and returns the content as markdown.
func (f *Fetcher) Fetch(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "", ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrBadScheme
	}

	// SSRF protection: resolve hostname and check for private IPs.
	if !f.allowPrivate {
		if err := checkSSRF(u.Hostname()); err != nil {
			return "", err
		}
	}

	resp, err := f.client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("ingest: fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: %d %s", ErrBadStatus, resp.StatusCode, resp.Status)
	}

	limited := io.LimitReader(resp.Body, maxBodySize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("ingest: read body: %w", err)
	}
	if len(body) > maxBodySize {
		return "", ErrBodyTooLong
	}

	md, err := htmltomarkdown.ConvertString(string(body))
	if err != nil {
		return "", fmt.Errorf("ingest: html-to-markdown: %w", err)
	}

	return strings.TrimSpace(md), nil
}

// checkSSRF rejects private, loopback, and link-local addresses.
func checkSSRF(hostname string) error {
	ips, err := net.LookupHost(hostname)
	if err != nil {
		// If DNS fails, check if hostname is an IP literal.
		ip := net.ParseIP(hostname)
		if ip == nil {
			return fmt.Errorf("ingest: DNS lookup failed: %w", err)
		}
		ips = []string{ip.String()}
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return ErrSSRF
		}
	}

	return nil
}
