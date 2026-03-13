# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| v0.1.x  | Yes       |

## Reporting a vulnerability

**Please do not open a public issue for security vulnerabilities.**

Use [GitHub Security Advisories](https://github.com/mipsou/mcp-biblium/security/advisories/new) to report privately. Include:

- Description of the vulnerability
- Steps to reproduce
- Impact assessment (if possible)

Reports will be triaged within 72 hours. Once confirmed, a fix will be prioritized and released as a patch version.

## Scope

Biblium runs locally via stdio — it has no network listener and no authentication layer. The main attack surface is:

- Path traversal in collection/document names (mitigated by `internal/safepath`)
- Malicious content in fetched URLs (`suggest_url` / `approve_url`)
- SQLite injection via pending URL storage

## Disclosure policy

- Vulnerabilities will be fixed before public disclosure
- Credit will be given to reporters (unless they prefer anonymity)
