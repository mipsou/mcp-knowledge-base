# Changelog

All notable changes to Biblium will be documented in this file.

## [v0.1.0] — 2026-03-13

First public release. Complete rewrite from Node.js to Go.

### Added
- 9 MCP tools: `create_collection`, `list_collections`, `add_document`, `list_documents`, `read_document`, `search`, `suggest_url`, `approve_url`, `list_pending`
- BM25 full-text search (in-memory index, rebuilt on startup)
- SQLite persistence for pending URLs (modernc.org/sqlite, pure Go)
- URL ingestion with approval workflow (fetch + convert to markdown)
- Path traversal protection (`internal/safepath`)
- Bilingual README (EN/FR) with Mermaid architecture diagrams
- Cross-compiled binaries: Linux (amd64/arm64), macOS (arm64), Windows (amd64), FreeBSD (amd64), OpenBSD (amd64)
- CI: Go test + build + golangci-lint + gosec
- SECURITY.md with GitHub Security Advisories
- EUPL-1.2 license (EN + FR)

### Known issues
- NetBSD: cross-compilation fails — modernc.org/sqlite missing NetBSD support ([#23](https://github.com/mipsou/mcp-biblium/issues/23))

[v0.1.0]: https://github.com/mipsou/mcp-biblium/releases/tag/v0.1.0
