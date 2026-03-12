# Biblium

[English](#english) | [Fran\u00e7ais](#français)

---

## English

MCP server for managing knowledge collections with full-text search.

Pure Go, zero CGO, single binary. Uses SQLite (via modernc.org/sqlite) and BM25 ranking.

### Features

- **Collections** \u2014 organize documents into named groups
- **BM25 search** \u2014 full-text search across all collections
- **URL ingestion** \u2014 suggest URLs, approve them, auto-fetch as markdown
- **SQLite persistence** \u2014 pending URLs stored in WAL-mode SQLite
- **MCP protocol** \u2014 stdio transport, works with Claude Desktop / Claude Code

### MCP Tools

| Tool | Description |
|------|-------------|
| `create_collection` | Create a new collection |
| `list_collections` | List all collections |
| `add_document` | Add a document to a collection |
| `list_documents` | List documents in a collection |
| `read_document` | Read a document |
| `search` | Search across all collections |
| `suggest_url` | Suggest a URL for ingestion (pending approval) |
| `approve_url` | Approve a pending URL |
| `list_pending` | List pending URL suggestions |

### Build

```bash
go build -o biblium ./cmd/biblium
```

Cross-compile (no CGO required):

```bash
GOOS=linux GOARCH=amd64 go build -o biblium ./cmd/biblium
```

### Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BIBLIUM_DATA_DIR` | `~/biblium_data` | Data storage directory |
| `BIBLIUM_SEARCH_BACKEND` | `bm25` | Search backend (`bm25` or `ollama`) |
| `BIBLIUM_LOG_LEVEL` | `info` | Log level |

### Usage with Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "biblium": {
      "command": "/path/to/biblium",
      "env": {
        "BIBLIUM_DATA_DIR": "/path/to/data"
      }
    }
  }
}
```

### License

EUPL-1.2-or-later \u2014 [Full text](https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12)

---

## Fran\u00e7ais

Serveur MCP pour g\u00e9rer des collections de connaissances avec recherche plein texte.

Go pur, z\u00e9ro CGO, binaire unique. Utilise SQLite (via modernc.org/sqlite) et classement BM25.

### Fonctionnalit\u00e9s

- **Collections** \u2014 organiser les documents en groupes nomm\u00e9s
- **Recherche BM25** \u2014 recherche plein texte sur toutes les collections
- **Ingestion d'URL** \u2014 proposer des URLs, les approuver, r\u00e9cup\u00e9ration auto en markdown
- **Persistance SQLite** \u2014 URLs en attente stock\u00e9es en SQLite mode WAL
- **Protocole MCP** \u2014 transport stdio, compatible Claude Desktop / Claude Code

### Outils MCP

| Outil | Description |
|-------|-------------|
| `create_collection` | Cr\u00e9er une nouvelle collection |
| `list_collections` | Lister toutes les collections |
| `add_document` | Ajouter un document \u00e0 une collection |
| `list_documents` | Lister les documents d'une collection |
| `read_document` | Lire un document |
| `search` | Rechercher dans toutes les collections |
| `suggest_url` | Proposer une URL \u00e0 ing\u00e9rer (approbation requise) |
| `approve_url` | Approuver une URL en attente |
| `list_pending` | Lister les URLs en attente |

### Compilation

```bash
go build -o biblium ./cmd/biblium
```

Cross-compilation (aucun CGO requis) :

```bash
GOOS=linux GOARCH=amd64 go build -o biblium ./cmd/biblium
```

### Configuration

Variables d'environnement :

| Variable | D\u00e9faut | Description |
|----------|--------|-------------|
| `BIBLIUM_DATA_DIR` | `~/biblium_data` | R\u00e9pertoire de stockage |
| `BIBLIUM_SEARCH_BACKEND` | `bm25` | Moteur de recherche (`bm25` ou `ollama`) |
| `BIBLIUM_LOG_LEVEL` | `info` | Niveau de log |

### Utilisation avec Claude Desktop

Ajouter dans `claude_desktop_config.json` :

```json
{
  "mcpServers": {
    "biblium": {
      "command": "/chemin/vers/biblium",
      "env": {
        "BIBLIUM_DATA_DIR": "/chemin/vers/data"
      }
    }
  }
}
```

### Licence

EUPL-1.2-ou-ult\u00e9rieure \u2014 [Texte int\u00e9gral](https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12)
