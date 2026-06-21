# Google Keep Clone

A research & development project to create a Google Keep clone that mirrors the official Google Keep REST API. Built with Goa (Go) for the backend and Next.js 15 for the frontend.

## Project Structure

```
├── backend/
│   ├── design/              # Goa DSL API design files
│   │   ├── design.go        # API metadata & server config
│   │   ├── types.go         # Data types (Note, Section, ListItem, etc.)
│   │   ├── notes.go         # Notes service design
│   │   ├── permissions.go   # Permissions service design
│   │   ├── media.go         # Media service design
│   │   └── labels.go        # Labels service design
│   ├── gen/                 # Generated Goa code (do not edit)
│   ├── api/                 # Business logic implementation
│   │   ├── notes.go
│   │   ├── permissions.go
│   │   ├── media.go
│   │   └── labels.go
│   ├── store/               # Database layer
│   │   ├── db.go            # PostgreSQL connection pool
│   │   ├── notes.go         # Note CRUD queries
│   │   ├── labels.go        # Label queries
│   │   ├── permissions.go   # Permission queries
│   │   ├── migrations/      # SQL migration files
│   │   └── migrate.go       # Migration runner
│   ├── cmd/keep_server/     # Server entrypoint
│   │   ├── main.go
│   │   └── http.go
│   ├── Dockerfile
│   └── go.mod
├── frontend/                # Next.js 15 app (placeholder)
├── k8s/                     # Kubernetes manifests
├── .github/workflows/       # CI pipeline
├── docker-compose.yml
└── README.md
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/notes` | Create a note |
| `GET` | `/v1/notes` | List notes (with pagination & filtering) |
| `GET` | `/v1/notes/{id}` | Get a note |
| `PATCH` | `/v1/notes/{id}` | Update a note |
| `DELETE` | `/v1/notes/{id}` | Delete a note |
| `POST` | `/v1/notes/{id}:pin` | Pin a note |
| `POST` | `/v1/notes/{id}:unpin` | Unpin a note |
| `POST` | `/v1/notes/{id}:archive` | Archive a note |
| `POST` | `/v1/notes/{id}:unarchive` | Unarchive a note |
| `POST` | `/v1/notes/{id}:trash` | Trash a note |
| `POST` | `/v1/notes/{id}:restore` | Restore a trashed note |
| `POST` | `/v1/notes/{noteId}/permissions:batchCreate` | Batch create permissions |
| `POST` | `/v1/notes/{noteId}/permissions:batchDelete` | Batch delete permissions |
| `GET` | `/v1/notes/{noteId}/attachments/{attachmentId}` | Download attachment |
| `GET` | `/v1/labels` | List labels |
| `POST` | `/v1/labels` | Create a label |
| `DELETE` | `/v1/labels/{id}` | Delete a label |

## Quick Start

```bash
# Start everything with Docker Compose
docker compose up

# Or run the backend directly
cd backend
DATABASE_URL="postgres://keep:keep@localhost:5432/keep?sslmode=disable" go run ./cmd/keep_server
```

## Development Workflow

1. **Modify the API design** in `backend/design/`
2. **Regenerate code**: `goa gen github.com/flaccid/google-keep-clone/backend/design`
3. **Update business logic** in `backend/api/`
4. **Build & run**: `go run ./cmd/keep_server`

The `gen/` directory is auto-generated — never edit it directly.
