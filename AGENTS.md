# AGENTS.md

## Project structure

Monorepo with a Go backend and Next.js 16 frontend.

```
backend/          # Go module github.com/flaccid/google-keep-clone/backend
  design/         # Goa DSL API design files — edit these first when adding endpoints
  gen/            # Auto-generated Goa code — NEVER edit manually
  api/            # Business logic implementation matching gen/ service interfaces
  store/          # PostgreSQL data layer (pgx/v5 pool, golang-migrate migrations)
  cmd/keep_server/  # Server entrypoint (main.go, http.go)
  migrations/     # SQL migration files in store/migrations/ (embedded)
frontend/         # Next.js 16 (App Router, TypeScript, Tailwind v4)
  src/components/ # React components
  src/lib/        # API client (api.ts), TypeScript types (types.ts)
  src/app/        # App Router pages
k8s/              # Kubernetes manifests per component
```

## Critical workflows

### Goa codegen (backend)
1. Edit DSL in `backend/design/*.go`
2. Run `make gen` in `backend/` — runs `goa gen` then copies `openapi3.yaml` to `cmd/keep_server/`
3. Update business logic in `backend/api/*.go` to match new generated interfaces
4. Build with `go build ./...` in `backend/`

The `gen/` directory is the compiled result of the Goa design — never edit manually.

### Testing
- Backend store tests require a running PostgreSQL: `go test -count=1 ./store/...`
- CI runs tests with `DATABASE_URL=postgres://keep:keep@localhost:5432/keep?sslmode=disable`
- Frontend: `npm run build` (includes TypeScript check)

### Docker
Two separate Dockerfiles, not a monorepo root one:
- `backend/Dockerfile` — multi-stage Go build
- `frontend/Dockerfile` — accepts `ARG API_UPSTREAM_URL=http://backend:8080` (build-time only, baked into Next.js rewrites)

## Key architecture facts

### Backend
- **Go 1.25.0** required (Goa v3.28.0 dependency)
- Route params accessed via `mux.Vars(r)` (Goa muxer), NOT `r.PathValue`
- DB auto-migrates on startup via `store.RunMigrations()` — no manual migration step
- Attachment upload is a standalone `net/http` handler registered on the Goa mux at `POST /v1/notes/{noteId}/attachments` — not a Goa-generated endpoint
- OpenAPI spec at `/openapi` (Swagger UI) and `/openapi.yaml` — embedded via `//go:embed openapi3.yaml`
- Notes store uses `ILIKE '%term%'` search on title and body_text via `search` query param
- Notes list accepts `filter` param (e.g. `filter=trashed=true`, `filter=archived=true`) — mapped to `notes.trashed` / `notes.archived` columns in the store
- `bodyFields()` in `api/notes.go` extracts body type/text/list items from `*notes.Section` for both Create and Update
- Server binds to `0.0.0.0:8080` (flag `-host=0.0.0.0`), NOT `localhost`
- Color values are `DEFAULT|RED|ORANGE|YELLOW|GREEN|TEAL|BLUE|CERULEAN|PURPLE|PINK|BROWN|GRAY|THEME_*` — frontend was broken before because DARK_BLUE was sent instead of CERULEAN

### Frontend
- **Next.js 16** — read `node_modules/next/dist/docs/` before writing code, APIs may differ from training data
- **Tailwind v4** — configured via `@import "tailwindcss"` and `@theme {}` block in `globals.css` (NO tailwind.config.ts or postcss.config.ts)
- **Dark mode**: class-based via `@custom-variant dark (&:where(.dark, .dark *))` in CSS, toggled by `ThemeProvider` which adds/removes `.dark` on `<html>` and persists to localStorage
- **API proxying**: Next.js rewrites proxy `/v1/*`, `/openapi`, `/openapi.yaml` to backend — no CORS needed
- **Sidebar**: collapsed = icons-only narrow 68px strip; hover when collapsed temporarily expands with overlay shadow; hamburger toggles expanded/collapsed
- **Color palette**: click-to-toggle (not CSS hover), shows two sections: Themes (gradient swatches) and Colors (solid circles)
- **Note labels**: `Note.labels` contains display name strings, NOT resource names. Label filter page must fetch all labels and match by displayName to filter correctly
- **NoteEditor label picker**: dropdown with checkboxes, labels sent as `labels[]` in create/update payloads

### K8s
- Namespace: `google-keep-clone`
- Deployed via `kubectl apply -f k8s/`
- Postgres uses StatefulSet with PVC; `wait-for-postgres` init container on backend
- OAuth2 proxy (Google provider) deployed as separate deployment with two ingress resources
- Ingress hostname: `keep.fordham.id.au`
