# Architecture Rewrite — Design Document

> Date: 2026-03-08
> Branch: dev-rewrite
> License: EUPL-1.2-or-later
> Status: DRAFT — pending approval

---

## 1. Executive Summary

This document describes the complete redesign of the legacy "Knowledge Base MCP
Server" project. The rewrite follows a clean-room approach: no code, naming, or
architecture is reused from the legacy TypeScript/Node.js implementation.

The new project adopts Go as its sole implementation language, targeting a single
statically-linked binary with file-based storage, zero external runtime
dependencies, and a pluggable search backend (BM25 by default, Ollama optional).

Design philosophy: PicoClaw/smollBSD-level simplicity. Target: under 4000 lines
of Go for the core.

---

## 2. Legacy Repository Assessment

### 2.1 Origin and Provenance

The current repository `mipsou/mcp-knowledge-base` (branch `dev`) is derived
from [jeanibarz/knowledge-base-mcp-server](https://github.com/jeanibarz/knowledge-base-mcp-server).
Evidence:

- README.md contains Smithery badge `@jeanibarz/knowledge-base-mcp-server`
- README.md contains Glama badge linking to `@jeanibarz`
- smithery.yaml configuration references the original project identity
- The `package.json` still shows `"license": "UNLICENSED"` while `LICENSE` is
  EUPL-1.2 (inconsistency)

The repository is not declared as a GitHub fork (`fork: false` in API metadata).

### 2.2 Architecture Assessment

| Aspect | Assessment |
|--------|------------|
| Language | TypeScript/Node.js (ES modules) |
| Entry point | `src/index.ts` → `KnowledgeBaseServer` |
| MCP transport | stdio via `@modelcontextprotocol/sdk` |
| Vector store | FAISS via `faiss-node` + LangChain |
| Embeddings | 4 providers: HuggingFace, OpenAI, Ollama, Xenova |
| URL management | JSON file (`pending_urls.json`) |
| Text splitting | LangChain `MarkdownTextSplitter` |
| HTTP client | Axios |
| HTML→MD | Turndown |
| Logging | Custom stderr logger |
| Tests | Jest (3 test files) |
| Container | Docker multi-stage (node:lts-alpine) |
| Dependencies (direct) | 11 production, 6 dev |

### 2.3 Dependency Tree Risk

| Dependency | Risk Level | Issue |
|------------|-----------|-------|
| `faiss-node: "latest"` | **CRITICAL** | Pinned to `latest` — supply chain attack vector |
| `langchain ^0.3.15` | HIGH | Massive framework (800+ transitive deps) |
| `@langchain/community ^0.3.29` | HIGH | Same ecosystem |
| `@langchain/core ^0.3.39` | HIGH | Same ecosystem |
| `@langchain/ollama ^0.2.3` | MEDIUM | LangChain sub-package |
| `@xenova/transformers ^2.17.2` | MEDIUM | Large ML runtime |
| `axios ^1.7.9` | LOW | Well-maintained HTTP client |
| `pickleparser ^0.2.1` | MEDIUM | Pickle deserialization (security risk) |
| `turndown ^7.2.2` | LOW | HTML to Markdown |
| `@huggingface/inference ^4.x` | LOW | API client |
| `@modelcontextprotocol/sdk ^1.0.4` | LOW | MCP protocol SDK |

Total transitive dependencies via `package-lock.json`: **7452 lines** — massive
supply chain surface.

### 2.4 Source Files (7 TypeScript modules)

| File | Lines | Purpose |
|------|-------|---------|
| `index.ts` | 8 | Entry point |
| `KnowledgeBaseServer.ts` | 240 | MCP server + tool handlers |
| `FaissIndexManager.ts` | 280 | FAISS index lifecycle |
| `UrlManager.ts` | 130 | URL suggest/approve/reject/fetch |
| `XenovaEmbeddings.ts` | 40 | Local embeddings wrapper |
| `config.ts` | 25 | Environment variable config |
| `logger.ts` | 60 | Stderr + file logger |
| `utils.ts` | 40 | SHA256 + recursive file listing |

**Total: ~823 lines of application code.**

---

## 3. Compliance and Plagiarism Risk Analysis

### 3.1 Provenance Issues

| Risk | Severity | Detail |
|------|----------|--------|
| Undeclared fork | HIGH | Code originates from jeanibarz project without attribution |
| License inconsistency | HIGH | `package.json` says UNLICENSED, LICENSE file says EUPL-1.2 |
| Original author badges | MEDIUM | README still shows jeanibarz Smithery/Glama badges |
| Smithery config | MEDIUM | `smithery.yaml` references original project |
| AI-generated code | LOW | Commits indicate Claude co-authorship (properly attributed) |

### 3.2 Clean-Room Rewrite Obligations

To eliminate all plagiarism risk, the rewrite MUST NOT:
- Copy any source code from the legacy implementation
- Reuse file names, class names, function names, or variable names
- Follow the same module decomposition
- Use the same configuration key names
- Reference the original project identity in any form

The rewrite MUST:
- Design architecture from requirements only (functional scope)
- Use entirely different naming conventions
- Use a different language (Go instead of TypeScript)
- Document the clean-room process

---

## 4. Security Risk Analysis (Legacy)

| Vulnerability | Severity | Detail |
|---------------|----------|--------|
| Path traversal | CRITICAL | `knowledge_base_name` parameter used directly in `path.join()` without validation — attacker can read arbitrary files |
| SSRF | HIGH | `add_url`/`approve_url` fetch any URL without allowlist — server can be used to probe internal network |
| Supply chain | HIGH | `faiss-node: "latest"` allows any version to be installed |
| Pickle deserialization | HIGH | `pickleparser` dependency enables arbitrary code execution via crafted pickle data |
| No input validation | MEDIUM | `args: any` with minimal type checking throughout |
| No audit logging | MEDIUM | No record of who requested what or when |
| Secrets in env vars | LOW | API keys passed via environment (standard but no validation) |
| No rate limiting | LOW | No abuse protection on MCP tool calls |

---

## 5. Reconstructed Functional Scope

From the legacy code analysis, the system provides these capabilities:

### Core Functions

1. **Corpus Management**: Organize documents into named collections (knowledge bases)
2. **Document Indexing**: Ingest text/markdown files, split into chunks, create searchable index
3. **Semantic Search**: Query the index with natural language, return ranked results
4. **URL Ingestion**: Fetch web pages, convert to markdown, add to corpus
5. **URL Approval Workflow**: Suggest → approve/reject → fetch → index pipeline
6. **Change Detection**: SHA256 hash-based incremental re-indexing

### MCP Interface (7 tools)

1. List available corpora
2. Search across corpora (with optional corpus filter and score threshold)
3. Suggest a URL for ingestion (pending approval)
4. List pending URL suggestions
5. Approve a pending URL (triggers fetch + index)
6. Reject a pending URL
7. Directly add a URL (skip approval)

### Non-Functional Requirements

- stdio transport (MCP standard)
- Configurable via environment variables
- Structured logging (stderr + optional file)
- Offline-capable (local embeddings)

---

## 6. New Project Name Proposals

The project needs a completely new identity. Requirements: no confusion with the
legacy project, suitable for open-source infrastructure, GitHub-compatible slug.

| # | Name | Slug | Description | Reasoning |
|---|------|------|-------------|-----------|
| 1 | **Cortex** | `cortex-mcp` | Knowledge retrieval cortex for AI agents | Brain metaphor for knowledge retrieval; short, memorable |
| 2 | **Recall** | `recall-mcp` | Fast knowledge recall server | Direct verb describing the function; simple |
| 3 | **Stacks** | `stacks-mcp` | Stackable knowledge collections for MCP | Metaphor for layered knowledge bases |
| 4 | **Silo** | `silo-mcp` | Isolated knowledge silos for AI | Storage metaphor; emphasizes separation of concerns |
| 5 | **Vault** | `vault-kb` | Knowledge vault with MCP access | Emphasizes security; but conflicts with HashiCorp Vault |
| 6 | **Mnemo** | `mnemo` | Memory and knowledge for AI agents | From Greek "mneme" (memory); unique, short |
| 7 | **Folio** | `folio-mcp` | Document folio server for MCP | Publishing/document metaphor; clear purpose |
| 8 | **Corpus** | `corpus-mcp` | Document corpus server for AI | Precise term for document collections |
| 9 | **Glean** | `glean-mcp` | Gather and retrieve knowledge | Verb meaning to collect/harvest; but Glean.com exists |
| 10 | **Lore** | `lore-mcp` | Knowledge lore server for MCP | Short, evocative; means accumulated knowledge/tradition |

### Recommended: **Lore** (`lore-mcp`)

Reasoning:
- 4 letters, extremely memorable
- "Lore" means accumulated knowledge, tradition, learning — perfect semantic fit
- No major trademark conflicts in the infrastructure space
- Clean GitHub slug: `lore-mcp`
- Distinct from any existing MCP server names
- Works in both English and French contexts

---

## 7. Technology Stack Evaluation

### 7.1 Decision: Go

Based on the PicoClaw/smollBSD design philosophy:

| Criterion | Go | Rust | Python |
|-----------|-----|------|--------|
| Single binary | Yes | Yes | No (needs interpreter) |
| Zero runtime deps | Yes | Yes | No (venv, pip) |
| Compile speed | Fast (~2s) | Slow (~30s+) | N/A |
| Memory footprint | <10MB | <5MB | >50MB |
| KISS factor | **Excellent** | Good (complex types) | Good (but runtime heavy) |
| MCP ecosystem | Growing | rmcp crate | Official SDK |
| ML/embeddings native | Weak (needs CGO) | Moderate | Excellent |
| TDD support | `go test` native | `cargo test` native | pytest |
| Cross-compilation | Trivial | Good | Fragile |
| Supply chain | `go.sum` lockfile | `cargo-audit` | pip audit |
| EUPL-1.2 compat | Yes (BSD stdlib) | Yes (MIT/Apache) | Yes (PSF) |
| Learning curve | Low | High | Low |
| PicoClaw alignment | **100%** | 0% | 0% |

### 7.2 Go Dependency Strategy

Target: **under 10 direct dependencies**.

| Need | Solution | Dep or Stdlib |
|------|----------|---------------|
| MCP protocol | `mcp-go` or hand-rolled JSON-RPC | 1 dep or stdlib |
| BM25 search | Pure Go implementation | stdlib (math, strings) |
| Ollama client | HTTP via `net/http` | stdlib |
| HTML→Markdown | `html-to-markdown` | 1 dep |
| SHA256 hashing | `crypto/sha256` | stdlib |
| File walking | `filepath.WalkDir` | stdlib |
| JSON handling | `encoding/json` | stdlib |
| Logging | `log/slog` (Go 1.21+) | stdlib |
| Config | Environment vars + simple TOML | 1 dep or stdlib |
| Testing | `testing` package | stdlib |
| CLI | `flag` package | stdlib |

**Estimated: 2-4 external dependencies.** Everything else is stdlib.

---

## 8. High-Level Architecture

```
┌─────────────────────────────────────────────┐
│                 lore-mcp                     │
│                                              │
│  ┌──────────┐   ┌──────────┐   ┌─────────┐ │
│  │  MCP      │──▶│  Router   │──▶│ Tools   │ │
│  │  Transport│   │           │   │         │ │
│  │  (stdio)  │◀──│           │◀──│         │ │
│  └──────────┘   └──────────┘   └────┬────┘ │
│                                      │      │
│                    ┌─────────────────┤      │
│                    ▼                 ▼      │
│              ┌──────────┐    ┌──────────┐  │
│              │  Corpus   │    │ Ingester │  │
│              │  Manager  │    │ (URL/    │  │
│              │           │    │  File)   │  │
│              └─────┬────┘    └─────┬────┘  │
│                    │               │        │
│                    ▼               ▼        │
│              ┌──────────────────────────┐  │
│              │       Searcher           │  │
│              │  (interface)             │  │
│              ├──────────┬───────────────┤  │
│              │ BM25     │ Ollama        │  │
│              │ (default)│ (optional)    │  │
│              └──────────┴───────────────┘  │
│                         │                   │
│                         ▼                   │
│              ┌──────────────────────────┐  │
│              │     Store (file-based)   │  │
│              │  corpus_dir/             │  │
│              │    ├── docs/             │  │
│              │    ├── index.json        │  │
│              │    └── meta.json         │  │
│              └──────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Single `main.go` entry point** — ~20 lines
2. **5 packages max**: `mcp`, `corpus`, `search`, `ingest`, `config`
3. **Interfaces for extensibility**: `Searcher`, `Embedder` (future)
4. **File-based store**: JSON metadata + markdown documents, no database
5. **stdlib first**: only add dependencies when stdlib is truly insufficient

---

## 9. Security Architecture

### 9.1 Threat Model

| Threat | Mitigation |
|--------|------------|
| Path traversal via corpus name | Canonical path validation: `filepath.Abs()` + verify prefix |
| SSRF via URL ingestion | URL allowlist (configurable domains), no private IP ranges |
| Malicious document content | Input size limits, no code execution from documents |
| Supply chain attack | Minimal deps, `go.sum` verification, `govulncheck` in CI |
| Unauthorized access | MCP transport security (stdio = same-user only) |
| Denial of service | Request rate limiting, corpus size limits, timeout on HTTP fetches |
| Log injection | Structured logging with `slog`, no user input in format strings |

### 9.2 Secure File Access

```go
// Every file path MUST go through this before use
func SafePath(root, userInput string) (string, error) {
    abs := filepath.Join(root, filepath.Clean(userInput))
    abs, err := filepath.Abs(abs)
    if err != nil {
        return "", err
    }
    rootAbs, _ := filepath.Abs(root)
    if !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) {
        return "", ErrPathTraversal
    }
    return abs, nil
}
```

### 9.3 URL Validation

- Parse URL, verify scheme is `http` or `https`
- Resolve hostname, reject private/loopback IP ranges (10.x, 172.16-31.x, 192.168.x, 127.x, ::1)
- Configurable domain allowlist (optional)
- Timeout: 30 seconds max
- Response size limit: 10MB max

### 9.4 Audit Logging

Every tool invocation logged via `slog` with:
- Timestamp (RFC3339)
- Tool name
- Sanitized parameters (no secrets)
- Result status (success/error)
- Duration

---

## 10. Licensing Architecture

### 10.1 SPDX Strategy

Every source file MUST include:
```go
// SPDX-License-Identifier: EUPL-1.2-or-later
// Copyright (C) 2026 Mipsou <chpujol@gmail.com>
```

### 10.2 Dependency License Compatibility

EUPL-1.2 compatible licenses (from EUPL Appendix):
- MIT ✅
- BSD-2-Clause ✅
- BSD-3-Clause ✅
- Apache-2.0 ✅
- MPL-2.0 ✅
- GPL-2.0, GPL-3.0 ✅ (via compatibility clause)
- LGPL-2.1, LGPL-3.0 ✅

Incompatible:
- Proprietary ❌
- SSPL ❌
- Commons Clause ❌

### 10.3 CI License Verification

- `go-licenses` in CI pipeline to check all dependencies
- THIRD-PARTY-NOTICES file auto-generated
- License check MUST pass before merge

---

## 11. Repository Structure

```
lore-mcp/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml          # build + test + lint + security
│   │   └── release.yml     # goreleaser
│   └── CODEOWNERS
├── cmd/
│   └── lore/
│       └── main.go         # entry point (~20 lines)
├── internal/
│   ├── config/
│   │   └── config.go       # env var loading + validation
│   ├── corpus/
│   │   ├── corpus.go       # corpus management
│   │   └── corpus_test.go
│   ├── ingest/
│   │   ├── ingest.go       # URL fetch + HTML→MD
│   │   └── ingest_test.go
│   ├── mcp/
│   │   ├── server.go       # MCP protocol handler
│   │   ├── tools.go        # tool definitions + dispatch
│   │   └── server_test.go
│   ├── search/
│   │   ├── searcher.go     # Searcher interface
│   │   ├── bm25.go         # BM25 implementation
│   │   ├── bm25_test.go
│   │   ├── ollama.go       # Ollama embeddings (optional)
│   │   └── ollama_test.go
│   └── safepath/
│       ├── safepath.go     # path traversal protection
│       └── safepath_test.go
├── docs/
│   └── plans/
├── .goreleaser.yml
├── .gitignore
├── CHANGELOG.md
├── go.mod
├── go.sum
├── LICENSE                  # EUPL-1.2 full text
├── THIRD-PARTY-NOTICES
└── README.md
```

**Estimated total: ~15 Go files, <3000 lines of code.**

---

## 12. New Storage Format

### 12.1 Corpus Directory Layout

```
$LORE_DATA_DIR/
├── corpora/
│   ├── infra/
│   │   ├── meta.toml           # corpus metadata
│   │   ├── docs/
│   │   │   ├── caddy-reverse-proxy.md
│   │   │   └── podman-networking.md
│   │   └── index.json          # BM25 inverted index (auto-generated)
│   ├── podman/
│   │   ├── meta.toml
│   │   ├── docs/
│   │   └── index.json
│   └── golang/
│       ├── meta.toml
│       ├── docs/
│       └── index.json
├── pending.json                # URL approval queue
└── lore.toml                   # global config (optional)
```

### 12.2 meta.toml

```toml
# Corpus metadata
name = "infra"
description = "Infrastructure documentation"
created_at = "2026-03-08T10:00:00Z"
```

### 12.3 index.json (auto-generated BM25 index)

```json
{
  "version": 1,
  "doc_count": 42,
  "avg_doc_len": 256.3,
  "terms": {
    "podman": {"df": 12, "postings": [{"doc": 0, "tf": 3}, ...]},
    "container": {"df": 8, "postings": [...]}
  },
  "documents": [
    {"id": 0, "path": "docs/podman-networking.md", "hash": "sha256:abc...", "len": 312}
  ]
}
```

### 12.4 Migration Tool

A `lore migrate` subcommand will convert legacy knowledge_bases/ format:
- Scan directories for .md files
- Create `meta.toml` per directory
- Move files into `docs/` subdirectory
- Build fresh BM25 index

---

## 13. Branching and Git Governance

### 13.1 Branch Strategy

| Branch | Purpose | Protection |
|--------|---------|------------|
| `main` | Production releases | Full protection |
| `dev` | Legacy integration | Read-only (archive) |
| `dev-rewrite` | Rewrite development | PR required |
| `feature/*` | Feature branches | None |

### 13.2 Branch Protection Rules (main)

- Direct push: **forbidden**
- Pull request: **mandatory**
- Approvals required: **1 minimum**
- Stale approvals: **dismissed on new push**
- Required CI checks: build, test, lint, security, license
- Conversations: **must be resolved**
- Signed commits: **required**
- Linear history: **enforced** (squash or rebase only)
- Admin bypass: **disabled**

### 13.3 CODEOWNERS

```
# All Go source files require review
*.go @mipsou

# Security-sensitive files require explicit review
internal/safepath/ @mipsou
internal/config/ @mipsou

# CI/CD changes require review
.github/ @mipsou
```

### 13.4 Commit Policy

- Signed commits required (GPG or SSH)
- Conventional commits format: `type(scope): description`
- Types: feat, fix, security, refactor, test, docs, chore
- Co-authorship for AI-generated code:
  `Co-Authored-By: Claude <noreply@anthropic.com>`

---

## 14. CI/CD and Supply-Chain Security

### 14.1 CI Pipeline (GitHub Actions)

```yaml
# Triggered on: push to dev-rewrite, PRs to main
jobs:
  build:
    - go build ./...
  test:
    - go test -race -cover ./...
    - coverage threshold: 80%
  lint:
    - golangci-lint run
  security:
    - govulncheck ./...        # known vulnerability check
    - gosec ./...               # static analysis (SAST)
  license:
    - go-licenses check ./...   # dependency license audit
  release:
    - goreleaser (on tag push to main)
```

### 14.2 Supply Chain Protection

- `go.sum` committed and verified
- Dependabot enabled for Go modules
- `govulncheck` on every PR
- No `replace` directives in `go.mod` (except local dev)
- Reproducible builds via `goreleaser` with checksums

### 14.3 DAST

- Not applicable for stdio MCP server (no network listener)
- URL ingestion tested with known-bad URLs in integration tests

---

## 15. Migration Strategy

### Phase 1: Clean-Room Build (dev-rewrite branch)

Build the entire new system from scratch. Zero reference to legacy code during
implementation. Only the functional scope document (Section 5) is used.

### Phase 2: Data Migration

The `lore migrate` tool converts existing knowledge_bases/ directories.
Users run: `lore migrate --from /path/to/knowledge_bases --to /path/to/lore-data`

### Phase 3: MCP Config Update

Users update their MCP client config to point to the new `lore` binary instead
of `node build/index.js`.

### Phase 4: Archive Legacy

The `dev` branch is archived. The `main` branch is updated from `dev-rewrite`.

---

## 16. Full Roadmap

| Phase | Duration | Milestone |
|-------|----------|-----------|
| P0: Skeleton | 1 day | Empty Go project, CI, license headers, branch setup |
| P1: Core corpus | 2 days | Corpus CRUD, file store, path safety |
| P2: BM25 search | 2 days | Pure Go BM25 with indexing |
| P3: MCP server | 2 days | stdio transport, tool registration, JSON-RPC |
| P4: Tool handlers | 2 days | All 7 MCP tools implemented |
| P5: URL ingestion | 1 day | HTTP fetch, HTML→MD, approval workflow |
| P6: Ollama search | 1 day | Optional semantic search via Ollama API |
| P7: Migration | 1 day | Legacy data migration tool |
| P8: Hardening | 2 days | Security review, fuzzing, edge cases |
| P9: Release | 1 day | goreleaser, README, CHANGELOG |

**Total: ~15 days**

---

## 17. Nano-Task Backlog

### Convention

Each nano-task follows:
- **ID**: `LOR-XXXX`
- **Title**: imperative verb phrase
- **Purpose**: why this task exists
- **Inputs/Outputs**: what goes in, what comes out
- **Files**: which files are created or modified
- **Dependencies**: which tasks must complete first
- **Validation**: how to verify completion
- **Security check**: what security concern does this address
- **Compliance check**: what license/provenance concern
- **Definition of Done**: test passes, review approved, CI green

---

## 18. First 20 Priority Nano-Tasks

### LOR-0001: Create dev-rewrite branch from main

- **Purpose**: Establish clean branch for the rewrite
- **Inputs**: main branch
- **Outputs**: Empty dev-rewrite branch
- **Files**: none (branch creation only)
- **Dependencies**: none
- **Validation**: `git branch -a` shows dev-rewrite
- **Security**: branch created from known-good main
- **Compliance**: no legacy code carried over
- **DoD**: Branch exists, pushed to origin

### LOR-0002: Initialize Go module

- **Purpose**: Set up Go project structure
- **Inputs**: Project name decision (lore-mcp)
- **Outputs**: `go.mod` with module path
- **Files**: `go.mod`
- **Dependencies**: LOR-0001
- **Validation**: `go mod tidy` succeeds
- **Security**: No dependencies yet
- **Compliance**: Module path matches project identity
- **DoD**: `go.mod` exists, `go build ./...` succeeds (empty)

### LOR-0003: Add EUPL-1.2 LICENSE file

- **Purpose**: Establish project license
- **Inputs**: EUPL-1.2 full text
- **Outputs**: `LICENSE` file
- **Files**: `LICENSE`
- **Dependencies**: LOR-0001
- **Validation**: File contains complete EUPL-1.2 text
- **Security**: N/A
- **Compliance**: SPDX identifier matches file content
- **DoD**: LICENSE file committed with correct content

### LOR-0004: Create .gitignore for Go project

- **Purpose**: Prevent build artifacts and sensitive files from being committed
- **Inputs**: Go project conventions
- **Outputs**: `.gitignore`
- **Files**: `.gitignore`
- **Dependencies**: LOR-0001
- **Validation**: `git status` does not show build artifacts
- **Security**: Excludes `.env`, credential files
- **Compliance**: N/A
- **DoD**: .gitignore committed

### LOR-0005: Create minimal main.go entry point

- **Purpose**: Establish compilable entry point
- **Inputs**: Architecture design
- **Outputs**: `cmd/lore/main.go` (~20 lines)
- **Files**: `cmd/lore/main.go`
- **Dependencies**: LOR-0002
- **Validation**: `go build ./cmd/lore/` produces binary
- **Security**: SPDX header present
- **Compliance**: Copyright header, SPDX-License-Identifier
- **DoD**: Binary compiles, runs (exits cleanly), headers correct

### LOR-0006: Implement SafePath module with tests

- **Purpose**: Path traversal protection — security foundation
- **Inputs**: Root directory path, user-provided relative path
- **Outputs**: Validated absolute path or error
- **Files**: `internal/safepath/safepath.go`, `internal/safepath/safepath_test.go`
- **Dependencies**: LOR-0002
- **Validation**: Tests cover: normal paths, `../` traversal, symlinks, empty input, absolute input
- **Security**: **Primary security control** — all file access MUST go through this
- **Compliance**: SPDX headers
- **DoD**: 100% test coverage on safepath, all traversal attacks blocked

### LOR-0007: Implement config loader with tests

- **Purpose**: Load and validate configuration from environment variables
- **Inputs**: Environment variables (`LORE_DATA_DIR`, `LORE_SEARCH_BACKEND`, etc.)
- **Outputs**: Validated `Config` struct
- **Files**: `internal/config/config.go`, `internal/config/config_test.go`
- **Dependencies**: LOR-0002
- **Validation**: Tests cover: defaults, overrides, invalid values, missing required
- **Security**: No secrets logged, validate data dir exists and is writable
- **Compliance**: SPDX headers
- **DoD**: Config loads from env, defaults work, invalid input rejected

### LOR-0008: Implement structured logging with slog

- **Purpose**: Audit trail and debugging
- **Inputs**: Config (log level, log file path)
- **Outputs**: Configured `slog.Logger`
- **Files**: Integrated in `internal/config/config.go` (logger setup)
- **Dependencies**: LOR-0007
- **Validation**: Test log output format, test log levels
- **Security**: Structured logging prevents log injection
- **Compliance**: SPDX headers
- **DoD**: Logger produces structured JSON to stderr, optional file output

### LOR-0009: Write Corpus data types and store interface

- **Purpose**: Define the corpus data model (TDD: write tests first)
- **Inputs**: Storage format design (Section 12)
- **Outputs**: `Corpus` struct, `Store` interface
- **Files**: `internal/corpus/corpus.go`, `internal/corpus/corpus_test.go`
- **Dependencies**: LOR-0006, LOR-0007
- **Validation**: Tests define expected behavior of List, Get, Create
- **Security**: All paths go through SafePath
- **Compliance**: SPDX headers
- **DoD**: Tests written (RED), types compile

### LOR-0010: Implement file-based corpus store

- **Purpose**: Implement the Store interface with file-system backend
- **Inputs**: LOR-0009 interface and tests
- **Outputs**: Working file-based corpus CRUD
- **Files**: `internal/corpus/corpus.go` (implement), `internal/corpus/corpus_test.go` (GREEN)
- **Dependencies**: LOR-0009
- **Validation**: All LOR-0009 tests pass (GREEN)
- **Security**: SafePath used for all file operations, directory creation validated
- **Compliance**: SPDX headers
- **DoD**: All tests GREEN, corpus can be listed/created/read from filesystem

### LOR-0011: Write BM25 search interface and test cases

- **Purpose**: Define search contract (TDD: tests first)
- **Inputs**: Search requirements from functional scope
- **Outputs**: `Searcher` interface, `Result` type, test expectations
- **Files**: `internal/search/searcher.go`, `internal/search/bm25_test.go`
- **Dependencies**: LOR-0009
- **Validation**: Tests define: index documents, query, rank results, score threshold
- **Security**: Query input sanitized (max length, no control chars)
- **Compliance**: SPDX headers
- **DoD**: Tests written (RED), interfaces compile

### LOR-0012: Implement BM25 ranking algorithm

- **Purpose**: Pure Go BM25 search implementation
- **Inputs**: LOR-0011 interface and tests
- **Outputs**: Working BM25 search with TF-IDF scoring
- **Files**: `internal/search/bm25.go`
- **Dependencies**: LOR-0011
- **Validation**: All LOR-0011 tests pass (GREEN)
- **Security**: Input length limits enforced
- **Compliance**: Algorithm is public domain (no IP concerns), SPDX headers
- **DoD**: All search tests GREEN, ranking produces sensible results

### LOR-0013: Implement BM25 index persistence

- **Purpose**: Save/load BM25 index to JSON file
- **Inputs**: In-memory BM25 index
- **Outputs**: `index.json` file per corpus
- **Files**: `internal/search/bm25.go` (add Save/Load)
- **Dependencies**: LOR-0012
- **Validation**: Tests: save index, reload, query returns same results
- **Security**: Index file written to SafePath-validated location only
- **Compliance**: SPDX headers
- **DoD**: Index survives process restart, tests GREEN

### LOR-0014: Implement document change detection

- **Purpose**: Only re-index changed documents (SHA256 hash comparison)
- **Inputs**: Document files in corpus
- **Outputs**: List of changed/new/deleted documents
- **Files**: `internal/corpus/corpus.go` (add hash tracking)
- **Dependencies**: LOR-0010
- **Validation**: Tests: add file → detected, modify → detected, unchanged → skipped
- **Security**: Hash comparison prevents unnecessary processing
- **Compliance**: SPDX headers
- **DoD**: Tests GREEN, only changed docs trigger re-index

### LOR-0015: Write MCP server skeleton with tests

- **Purpose**: MCP JSON-RPC over stdio (TDD: tests first)
- **Inputs**: MCP protocol specification
- **Outputs**: Server that handles `initialize`, `tools/list`, `tools/call`
- **Files**: `internal/mcp/server.go`, `internal/mcp/server_test.go`
- **Dependencies**: LOR-0005
- **Validation**: Tests: valid initialize handshake, tool listing, unknown method error
- **Security**: Strict JSON-RPC validation, reject malformed requests
- **Compliance**: SPDX headers
- **DoD**: Tests written (RED), MCP types defined

### LOR-0016: Implement MCP stdio transport

- **Purpose**: Read/write JSON-RPC messages on stdin/stdout
- **Inputs**: LOR-0015 tests
- **Outputs**: Working stdio MCP transport
- **Files**: `internal/mcp/server.go`
- **Dependencies**: LOR-0015
- **Validation**: All LOR-0015 tests pass (GREEN)
- **Security**: Input size limits (max message 1MB), no unbounded reads
- **Compliance**: SPDX headers
- **DoD**: Tests GREEN, MCP handshake works via stdio

### LOR-0017: Implement list_corpora tool

- **Purpose**: MCP tool to list available corpora (replaces list_knowledge_bases)
- **Inputs**: Corpus store
- **Outputs**: JSON list of corpus names
- **Files**: `internal/mcp/tools.go`, `internal/mcp/tools_test.go`
- **Dependencies**: LOR-0010, LOR-0016
- **Validation**: Test: create 3 corpora, call tool, verify 3 returned
- **Security**: No user input in this tool (safe)
- **Compliance**: SPDX headers, tool name differs from legacy
- **DoD**: Test GREEN, tool callable via MCP protocol

### LOR-0018: Implement search tool

- **Purpose**: MCP tool for searching across corpora
- **Inputs**: Query string, optional corpus name, optional threshold
- **Outputs**: Ranked search results with scores
- **Files**: `internal/mcp/tools.go`, `internal/mcp/tools_test.go`
- **Dependencies**: LOR-0012, LOR-0016
- **Validation**: Test: index docs, search, verify ranked results returned
- **Security**: Query sanitized, corpus name validated via SafePath
- **Compliance**: SPDX headers, tool name differs from legacy
- **DoD**: Test GREEN, search returns ranked results via MCP

### LOR-0019: Write URL ingestion module with tests

- **Purpose**: Fetch URL, convert HTML to markdown, save to corpus (TDD)
- **Inputs**: URL string, target corpus name
- **Outputs**: Markdown file in corpus docs/ directory
- **Files**: `internal/ingest/ingest.go`, `internal/ingest/ingest_test.go`
- **Dependencies**: LOR-0010
- **Validation**: Tests with httptest server: fetch, convert, verify markdown output
- **Security**: URL validation (scheme, no private IPs), timeout, size limit
- **Compliance**: SPDX headers, html-to-markdown dependency license check
- **DoD**: Tests GREEN, URL fetched and converted safely

### LOR-0020: Implement URL approval workflow

- **Purpose**: suggest → approve/reject pipeline for URL ingestion
- **Inputs**: URL, corpus name, reason
- **Outputs**: Pending entry in `pending.json`, approval triggers fetch+index
- **Files**: `internal/ingest/ingest.go` (add approval), `internal/ingest/ingest_test.go`
- **Dependencies**: LOR-0019
- **Validation**: Tests: suggest → list pending → approve → verify indexed; reject → verify removed
- **Security**: Pending file written atomically, IDs are UUIDs
- **Compliance**: SPDX headers
- **DoD**: Tests GREEN, full approval workflow works

---

## 19. Risks and Open Questions

### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| BM25 insufficient for semantic queries | MEDIUM | MEDIUM | Ollama backend available as upgrade path |
| Go MCP SDK immature | LOW | LOW | Protocol is simple JSON-RPC, can hand-roll if needed |
| HTML-to-markdown conversion quality | LOW | MEDIUM | Test with diverse real-world pages |
| Index size with large corpora | LOW | LOW | BM25 index is compact; add pagination if needed |

### Open Questions

1. ~~**MCP SDK choice**: Use `mark3labs/mcp-go` or implement minimal JSON-RPC handler?~~
   **Resolved**: `mark3labs/mcp-go` v0.45.0

2. **Text chunking strategy**: Fixed size (legacy: 1000 chars) or paragraph-based?
   Decision needed during LOR-0012.

3. **Concurrent access**: Single writer or need locking for index updates?
   Decision needed during LOR-0013.

4. **Container strategy**: Scratch image with static binary or distroless?
   Decision needed during P9.

5. **Data directory topology**: Where do corpora live relative to projects?
   Three modes to evaluate:
   - **Centralized**: Single global `LORE_DATA_DIR` shared by all projects (current default)
   - **Decentralized**: Each project embeds its own KB (`./lore-data/`), versioned with git
   - **Hybrid**: Global KB + per-project override via `.mcp.json` env `LORE_DATA_DIR`
   Decision needed before P7 (migration). Impacts: multi-datadir support, index merging,
   cross-project search, `.gitignore` patterns.

---

## Appendix A: Name Comparison with Legacy

| Aspect | Legacy | Rewrite |
|--------|--------|---------|
| Project name | Knowledge Base MCP Server | Lore |
| Binary name | `node build/index.js` | `lore` |
| KB collections | knowledge_bases | corpora |
| Tool: list | list_knowledge_bases | list_corpora |
| Tool: search | retrieve_knowledge | search |
| Tool: suggest | suggest_url | suggest |
| Tool: list pending | list_pending_urls | pending |
| Tool: approve | approve_url | approve |
| Tool: reject | reject_url | reject |
| Tool: direct add | add_url | ingest |
| Config root dir | KNOWLEDGE_BASES_ROOT_DIR | LORE_DATA_DIR |
| Config index path | FAISS_INDEX_PATH | (auto, inside corpus) |
| Config provider | EMBEDDING_PROVIDER | LORE_SEARCH_BACKEND |

Every name, every variable, every concept is renamed to eliminate any
resemblance to the legacy codebase.
