# 📝 Google Keep Clone

> A research & development project mirroring the official Google Keep REST API — built with Goa (Go), Next.js 16, and PostgreSQL.

[![CI](https://github.com/flaccid/google-keep-clone/actions/workflows/ci.yml/badge.svg)](https://github.com/flaccid/google-keep-clone/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![Node Version](https://img.shields.io/badge/Node-22-339933?logo=nodedotjs)](https://nodejs.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=nextdotjs)](https://nextjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## ✨ Features

| Feature | Status |
|---------|--------|
| 📝 Create, edit, delete notes | ✅ |
| 📋 Checklist / list notes | ✅ |
| 🎨 Solid color backgrounds (10 colours) | ✅ |
| 🌈 Gradient theme backgrounds (8 themes) | ✅ |
| 📌 Pin / unpin notes | ✅ |
| 🗂️ Archive / unarchive notes | ✅ |
| 🗑️ Trash / restore / permanent delete | ✅ |
| 🏷️ Label management (CRUD + rename) | ✅ |
| 🔍 Full-text search (ILIKE) | ✅ |
| 🌙 Dark mode (persisted to localStorage) | ✅ |
| 📎 Attachment upload & download | ✅ |
| 🔗 Share permissions (batch create/delete) | ✅ |
| 🧭 Sidebar navigation (collapsible mini mode) | ✅ |
| 🏠 Kubernetes deployment (with OAuth2 proxy) | ✅ |
| 📖 OpenAPI 3.0 docs (Swagger UI) | ✅ |

---

## 🏗️ Tech Stack

**Backend**
- **Go 1.25** with **Goa v3.28** (design-first API framework)
- **PostgreSQL 16** with **pgx/v5** connection pool
- **golang-migrate/v4** for schema migrations
- **Goa CLUE** for structured logging & debugging
- **go:embed** for OpenAPI spec bundling

**Frontend**
- **Next.js 16** (App Router, standalone output)
- **React 19** with **TypeScript 5**
- **Tailwind CSS v4** (no config file — `@import "tailwindcss"` in `globals.css`)
- **Lucide React** icons
- **clsx** for conditional classNames

**Infrastructure**
- Docker Compose (local dev)
- Kubernetes manifests (production-like deployment)
- OAuth2 Proxy with Google provider
- GitHub Actions CI (lint → build → test → Docker push)

---

## 🗺️ Project Structure

```
.
├── backend/                       # Go module: github.com/flaccid/google-keep-clone/backend
│   ├── design/                    # Goa DSL — edit these first
│   │   ├── design.go              #   API metadata & server config
│   │   ├── types.go               #   Data types (Note, Section, ListItem, etc.)
│   │   ├── notes.go               #   Notes service (14 methods)
│   │   ├── labels.go              #   Labels service (CRUD + update)
│   │   ├── permissions.go         #   Permissions service (batch create/delete)
│   │   └── media.go               #   Media service (download)
│   ├── gen/                       # 🔒 Auto-generated Goa code — never edit
│   ├── api/                       # Business logic
│   │   ├── notes.go
│   │   ├── labels.go
│   │   ├── permissions.go
│   │   └── media.go
│   ├── store/                     # PostgreSQL data layer
│   │   ├── db.go                  #   pgx/v5 connection pool
│   │   ├── notes.go               #   Note CRUD, search, label join
│   │   ├── labels.go              #   Label CRUD
│   │   ├── permissions.go         #   Permission queries
│   │   ├── attachments.go         #   File-based attachment store
│   │   ├── migrations/            #   SQL migration files
│   │   ├── migrate.go             #   Auto-migration runner
│   │   ├── notes_test.go          #   Integration tests (17 tests)
│   │   ├── labels_test.go         #   Integration tests (3 tests)
│   │   └── permissions_test.go    #   Integration tests (5 tests)
│   ├── cmd/keep_server/           # Server entrypoint
│   │   ├── main.go                #   Flags, DB init, service wiring
│   │   ├── http.go                #   HTTP server, upload handler, OpenAPI handler
│   │   └── openapi3.yaml          #   📎 Embedded OpenAPI spec
│   ├── Makefile                   #   `make gen` → goa gen + spec copy
│   ├── Dockerfile                 #   Multi-stage Go build → alpine
│   ├── go.mod                     #   Go 1.25.0
│   └── go.sum
├── frontend/                      # Next.js 16 (App Router, TypeScript, Tailwind v4)
│   ├── src/
│   │   ├── app/                   # App Router pages
│   │   │   ├── page.tsx           #   Home / note grid
│   │   │   ├── archive/page.tsx   #   Archived notes
│   │   │   ├── trash/page.tsx     #   Trashed notes
│   │   │   ├── reminders/page.tsx #   Reminders (placeholder)
│   │   │   └── labels/
│   │   │       ├── page.tsx       #   Label manager (inline rename/delete)
│   │   │       └── [id]/page.tsx  #   Notes filtered by label
│   │   ├── components/            # React components
│   │   │   ├── Header.tsx         #   Fixed header with lightbulb icon
│   │   │   ├── Sidebar.tsx        #   Collapsible sidebar (mini mode)
│   │   │   ├── NoteEditor.tsx     #   Inline "Take a note..." + label picker
│   │   │   ├── NoteCard.tsx       #   Note display card + color/theme palette
│   │   │   ├── NoteModal.tsx      #   Full-note overlay editor
│   │   │   ├── Palette.tsx        #   Shared color/theme picker
│   │   │   └── ThemeProvider.tsx  #   Dark mode context + localStorage
│   │   └── lib/
│   │       ├── api.ts             #   Full API client (notes, labels, permissions)
│   │       └── types.ts           #   TypeScript interfaces
│   ├── next.config.ts             #   Rewrites: /v1/* → backend
│   ├── Dockerfile                 #   Multi-stage Next.js standalone
│   └── package.json               #   Next.js 16, React 19, Tailwind v4
├── k8s/                           # Kubernetes manifests
│   ├── namespace.yaml
│   ├── ingress.yaml               #   Two ingresses (auth + oauth2)
│   ├── backend/                   #   Backend deployment, service, HPA
│   ├── frontend/                  #   Frontend deployment, service
│   ├── postgres/                  #   StatefulSet + service (PVC)
│   └── oauth2-proxy/              #   OAuth2 proxy (Google provider)
├── docker-compose.yml             # postgres + backend + frontend
└── .github/workflows/ci.yml       # Lint → Build → Test → Docker push
```

---

## 🌐 API Endpoints

All endpoints are prefixed with `/v1` and proxied via Next.js rewrites (no CORS needed).

### Notes

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/notes` | Create a note (text or list) |
| `GET` | `/v1/notes` | List notes (pagination, `search`, `filter`) |
| `GET` | `/v1/notes/{id}` | Get a single note |
| `PATCH` | `/v1/notes/{id}` | Update a note (title, body, color, labels) |
| `DELETE` | `/v1/notes/{id}` | Permanently delete a note |
| `POST` | `/v1/notes/{id}:pin` | Pin a note |
| `POST` | `/v1/notes/{id}:unpin` | Unpin a note |
| `POST` | `/v1/notes/{id}:archive` | Archive a note |
| `POST` | `/v1/notes/{id}:unarchive` | Unarchive a note |
| `POST` | `/v1/notes/{id}:trash` | Trash a note |
| `POST` | `/v1/notes/{id}:restore` | Restore from trash |

### Labels

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/labels` | List all labels |
| `POST` | `/v1/labels` | Create a label |
| `PATCH` | `/v1/labels/{id}` | Rename a label |
| `DELETE` | `/v1/labels/{id}` | Delete a label |

### Permissions

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/notes/{noteId}/permissions:batchCreate` | Batch create share permissions |
| `POST` | `/v1/notes/{noteId}/permissions:batchDelete` | Batch delete share permissions |

### Media

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/notes/{noteId}/attachments/{attachmentId}` | Download an attachment |

> **Note:** Attachment **upload** is at `POST /v1/notes/{noteId}/attachments` with `multipart/form-data` — a standalone `net/http` handler registered on the Goa mux (not Goa-generated).

### OpenAPI

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/openapi` | Swagger UI |
| `GET` | `/openapi.yaml` | Raw OpenAPI 3.0 spec |

**Query parameters for `GET /v1/notes`:**
- `pageSize` — Number of notes per page (default: 10)
- `pageToken` — Pagination cursor
- `search` — Full-text search (ILIKE on title + body_text)
- `filter` — e.g. `filter=trashed=true`, `filter=archived=true`

**Color values:** `DEFAULT | RED | ORANGE | YELLOW | GREEN | TEAL | BLUE | CERULEAN | PURPLE | PINK | BROWN | GRAY | THEME_SHORE | THEME_BLOOM | THEME_PLUM | THEME_NIGHT | THEME_BAMBOO | THEME_CANDY | THEME_SUNSET | THEME_OCEAN`

---

## 🚀 Quick Start

### Docker Compose

```bash
docker compose up
```

This starts three services:
1. **PostgreSQL 16** on port `5432`
2. **Backend** on port `8080` (auto-migrates DB on startup)
3. **Frontend** on port `3000`

Visit **http://localhost:3000** — all API calls are proxied from the Next.js dev server to the Go backend.

### Local development (without Docker)

```bash
# Terminal 1 — Database
docker run -d --name keep-pg \
  -e POSTGRES_USER=keep \
  -e POSTGRES_PASSWORD=keep \
  -e POSTGRES_DB=keep \
  -p 5432:5432 \
  postgres:16-alpine

# Terminal 2 — Backend
cd backend
DATABASE_URL="postgres://keep:keep@localhost:5432/keep?sslmode=disable" \
  go run ./cmd/keep_server -host=localhost

# Terminal 3 — Frontend
cd frontend
npm install
npm run dev
```

---

## 🧑‍💻 Development Workflow

### Backend — Goa codegen cycle

```bash
cd backend

# 1. Edit the DSL in design/*.go
# 2. Regenerate Goa code + copy OpenAPI spec
make gen

# 3. Update business logic in api/*.go
# 4. Build & verify
go build ./...
go vet ./...

# 5. Run store integration tests (requires PostgreSQL)
DATABASE_URL="postgres://keep:keep@localhost:5432/keep?sslmode=disable" \
  go test -count=1 ./store/... -v
```

**Rules:**
- `gen/` is **auto-generated** — never edit manually
- Route params use `mux.Vars(r)` (Goa muxer), not `r.PathValue`
- The DB **auto-migrates** on startup — no manual migration step needed
- After `make gen`, always sync `cmd/keep_server/openapi3.yaml` (already done by the Makefile)

### Frontend

```bash
cd frontend
npm install
npm run dev     # Development server on :3000
npm run build   # TypeScript check + production build
npm run lint    # ESLint
```

**Key details:**
- API calls use **relative paths** (`/v1/notes`, etc.) — CORS is not needed
- Next.js **rewrites** proxy `/v1/*`, `/openapi`, `/openapi.yaml` to the backend
- `API_UPSTREAM_URL` is a **build-time ARG** baked into rewrites, not a runtime env var
- Tailwind v4 is configured via `@import "tailwindcss"` + `@theme {}` in `globals.css` — no `tailwind.config.ts` or `postcss.config.ts`

---

## ☸️ Kubernetes Deployment

```bash
kubectl apply -f k8s/
```

Components (all in namespace `google-keep-clone`):

| Component | Type | Details |
|-----------|------|---------|
| PostgreSQL | StatefulSet + Service | PersistentVolumeClaim, init container wait |
| Backend | Deployment + Service + HPA | 100m CPU / 128Mi requests, auto-migration |
| Frontend | Deployment + Service | Standalone Next.js output |
| OAuth2 Proxy | Deployment + Service | Google OAuth, email whitelist |
| Ingress | 2 Ingress resources | Main (auth-protected) + `/oauth2` (unauthenticated) |

**Ingress hostname:** `keep.fordham.id.au` (Cloudflare tunnel to ingress-nginx)

To update images:
```bash
kubectl set image deployment/backend backend=flaccid/google-keep-clone-backend:latest
kubectl set image deployment/frontend frontend=flaccid/google-keep-clone-frontend:latest
```

---

## 🔄 CI/CD Pipeline

GitHub Actions runs on every push/PR to `main`:

1. **Lint** — `go vet ./...`
2. **Build & Test** — `go build ./...` + store integration tests (with PostgreSQL container)
3. **Frontend** — `npm ci` + `npm run build` (includes TypeScript check)
4. **Docker push** — On `main` only, pushes `flaccid/google-keep-clone-{backend,frontend}:latest` and `:{sha}`

---

## 🎨 UI Highlights

- **Sidebar:** Collapsed = icons-only 68px strip; hover temporarily expands with shadow overlay; hamburger toggles expanded/collapsed; main content margin adjusts accordingly
- **Color palette:** Click-to-toggle dropdown with two sections — **Themes** (gradient swatches) and **Colors** (solid circles)
- **Dark mode:** Toggle in settings dropdown, persists to localStorage, class-based via Tailwind `dark:` variant
- **Note editor:** Inline "Take a note..." with expand, text/list toggle, label picker, color/theme palette
- **Label manager:** Inline rename (pencil icon), delete with confirmation, sidebar "Edit labels" link
- **Labels on notes:** `Note.labels` contains display name strings (not resource names); label filter page fetches all labels and matches by displayName

---

## 🗄️ Database Schema

| Table | Purpose |
|-------|---------|
| `notes` | Core note data (title, body_text, color, pinned, archived, trashed) |
| `list_items` | Checklist items (linked to notes, ordered, with checked state) |
| `labels` | Label definitions (name, hex color) |
| `note_labels` | Many-to-many join between notes and labels |
| `permissions` | Note sharing (email, role, type) |
| `attachments` | File metadata (filename, size, mime type, linked to notes) |

Migrations live in `backend/store/migrations/` and are embedded into the binary via `//go:embed`.

---

## 🤝 Contributing

This is a personal R&D project, but issues and PRs are welcome! Feel free to open a discussion for major changes first.

---

## 📄 License

MIT
