# Plan: Rename lore → biblium (mcp-biblium)

> Date: 2026-03-12
> Branch: feature/rename-biblium (from dev)
> Status: PENDING APPROVAL

---

## 1. Scope

Rename projet "Lore" → "Biblium" (`mcp-biblium`). Touches:
- Go module path: `github.com/mipsou/lore-mcp` → `github.com/mipsou/mcp-biblium`
- Binary: `lore` → `biblium`
- Env vars: `LORE_*` → `BIBLIUM_*`
- Data dir default: `lore_data` → `biblium_data`
- Directory: `cmd/lore/` → `cmd/biblium/`
- GitHub repo: `mipsou/mcp-knowledge-base` → `mipsou/mcp-biblium`
- MCP server name: `lore` → `biblium`

---

## 2. Tasks (parallel where possible)

### Task A — Go source rename (agent 1)
1. Rename `cmd/lore/` → `cmd/biblium/`
2. Update `go.mod` module path
3. Replace all import paths (`lore-mcp` → `mcp-biblium`) in 6 Go files
4. Replace MCP server name `"lore"` → `"biblium"` in `server.go`
5. Replace log prefixes `"lore:"` → `"biblium:"` in `main.go`
6. Replace env vars `LORE_*` → `BIBLIUM_*` in `config.go` + `config_test.go`
7. Replace default data dir `lore_data` → `biblium_data` in `config.go`
8. Update `.gitignore` binary names
9. Run `go mod tidy` + `go test ./...` → must pass

### Task B — Config files (agent 2, parallel with A)
1. Update `.mcp.json`: server key, binary path, env vars
2. Update `.claude/settings.local.json`: tool names, build commands

### Task C — Documentation (agent 3, parallel with A+B)
1. Update `docs/plans/2026-03-08-lore-implementation-plan.md`
2. Update `docs/plans/2026-03-08-rewrite-design.md`
3. Update Appendix A name table

### Task D — GitHub repo rename (manual, after merge)
1. `gh repo rename mcp-biblium` on `mipsou/mcp-knowledge-base`
2. GitHub auto-redirects old URLs

### Task E — Memory + external refs (after D)
1. Update MEMORY.md (lore → biblium)
2. Update CLAUDE.md KB section if needed
3. Rebuild binary: `go build -o biblium.exe ./cmd/biblium/`

---

## 3. Files affected

| File | Changes |
|------|---------|
| `go.mod` | module path |
| `cmd/lore/main.go` → `cmd/biblium/main.go` | imports + log prefix |
| `internal/config/config.go` | env vars + default dir |
| `internal/config/config_test.go` | env var names |
| `internal/corpus/store.go` | import path |
| `internal/mcpserver/server.go` | import paths + server name |
| `internal/mcpserver/handlers_test.go` | import paths |
| `internal/mcpserver/server_test.go` | import paths |
| `internal/mcpserver/integration_test.go` | import paths |
| `.gitignore` | binary names |
| `.mcp.json` | server config |
| `.claude/settings.local.json` | tool names + commands |
| `docs/plans/*.md` (×2) | all "lore" references |

---

## 4. Validation

- `go test ./...` — 47 tests pass
- `go build -o biblium.exe ./cmd/biblium/` — builds clean
- Zero occurrences of "lore" in Go source (grep verify)
- MCP server responds as "biblium" via JSON-RPC test

---

## 5. Git strategy

- Branch: `feature/rename-biblium` from `dev`
- Single commit: `refactor: rename lore → biblium (mcp-biblium)`
- PR → `dev`, merge classique
- Repo rename after merge
