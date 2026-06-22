# AGENTS.md

## Project structure

```
backend/          # Go module github.com/flaccid/google-keep-clone/backend
  design/         # Goa DSL — edit these first when adding endpoints
  gen/            # Auto-generated Goa code — NEVER edit manually
  api/            # Business logic matching gen/ service interfaces
  store/          # PostgreSQL data layer (pgx/v5 + golang-migrate)
  cmd/keep_server/# Server entrypoint (main.go, http.go)
  migrations/     # SQL in store/migrations/, embedded via //go:embed
frontend/         # Next.js 16 (App Router, TypeScript, Tailwind v4)
  src/app/        # App Router pages (home, archive, trash, labels/*)
  src/components/ # React components
  src/lib/        # API client (api.ts), types (types.ts)
charts/           # Helm chart (recommended deploy method)
k8s/              # Raw K8s manifests (legacy, use chart instead)
docs/             # getting-started.md, api-reference.md
```

## Development commands

```bash
# Backend
cd backend && make gen          # Goa codegen + copy openapi3.yaml
cd backend && go build ./...    # Build verification
cd backend && go vet ./...      # Lint (CI lint job)
cd backend && DATABASE_URL=postgres://keep:keep@localhost:5432/keep?sslmode=disable go test -count=1 ./store/...

# Frontend
cd frontend && npm run dev      # Dev server :3000
cd frontend && npm run build    # TypeScript check + production build
cd frontend && npm run lint     # ESLint (CI frontend job)

# Docker
docker compose up               # postgres + backend + frontend
```

## Goa codegen workflow

1. Edit `backend/design/*.go` (DSL)
2. Run `make gen` in `backend/` — calls `goa gen` then copies `gen/http/openapi3.yaml` to `cmd/keep_server/openapi3.yaml`
3. Update `backend/api/*.go` to match new generated interfaces
4. `go build ./...` to verify

**Never edit `gen/` manually.** Route params use `mux.Vars(r)` (Goa muxer), not `r.PathValue`.

## Key backend quirks

- **DB auto-migrates** on startup via `store.RunMigrations()` — no manual migration step
- **Attachment upload** is a standalone `net/http` handler registered on the Goa mux at `POST /v1/notes/{noteId}/attachments` — not Goa-generated
- **Search**: `ILIKE '%term%'` on `title` and `body_text` via `search` query param. `%` and `_` escaped with `strings.ReplaceAll` before interpolation. Parameterized `$N` placeholders.
- **Filter**: whitelist-validated via `safeFilter()` — only `trashed`, `archived`, `pinned` with `true`/`false` values, joined by ` AND `
- **Color values**: `DEFAULT|RED|ORANGE|YELLOW|GREEN|TEAL|BLUE|CERULEAN|PURPLE|PINK|BROWN|GRAY|THEME_*` — both solids and themes stored in same `color` column
- **Body fields**: `bodyFields()` in `api/notes.go` extracts `type`/`text`/`listItems` from `*notes.Section`
- **Goa error sanitization**: custom `sanitizingFormatter` replaces fault messages with `"internal server error"` — DB internals never leak to clients
- **MIME allowlist**: upload validates 8 types (JPEG/PNG/GIF/WebP/SVG/PDF/text/octet-stream), others get 415
- Server binds `0.0.0.0:8080` (flag `-host=0.0.0.0`), `DATABASE_URL` env var required (no fallback)

## Key frontend quirks

- **Tailwind v4**: configured via `@import "tailwindcss"` + `@theme {}` in `globals.css` — NO `tailwind.config.ts` or `postcss.config.ts`
- **Dark mode**: class-based via `@custom-variant dark (&:where(.dark, .dark *))` in CSS, toggled by `ThemeProvider` which adds/removes `.dark` on `<html>` and persists to localStorage
- **No CORS**: Next.js rewrites proxy `/v1/*`, `/openapi`, `/openapi.yaml` to backend
- **`API_UPSTREAM_URL`** is a build-time ARG (Dockerfile), baked into rewrites — not runtime
- **Note.labels**: display name strings, NOT resource names. Label filter page fetches all labels and matches by `displayName`
- **Sidebar**: 68px mini when collapsed; hover temporarily expands with overlay shadow
- **Theme/color**: palette shows two sections — Themes (gradient swatches) and Colors (solid circles)

## Helm chart (recommended deploy)

Located at `charts/`. Values in `charts/values.yaml`. Install:

```bash
# Cookie secret must decode to 16, 24, or 32 bytes
COOKIE_SECRET=$(openssl rand -base64 24)
helm install google-keep-clone ./charts --set ingress.host=keep.example.com ...
```

### Chart-known gotchas

- **Postgres needs init container** to create `subPath: pgdata` with correct ownership. See `charts/templates/postgres/statefulset.yaml`: an `init-pgdata` container (busybox, runs as root) creates `/var/lib/postgresql/data/pgdata` and `chown 999:999` before postgres starts.
- **Backend init container** (`wait-for-postgres`) uses `postgres:16-alpine` and must specify `runAsUser: 999` because the pod has `runAsNonRoot: true` at the pod level.
- **OAuth2 cookie secret** must be a raw string of exactly 16, 24, or 32 bytes. Use `openssl rand -base64 24` (produces 32 chars = 32 bytes, no padding). Do NOT pipe through `base64` again.
- **Network policies**: `default-deny-all` blocks all ingress+egress. Explicit egress policies needed:
  - `allow-backend-egress-to-postgres` (backend → postgres:5432)
  - `allow-backend-egress-to-oauth2` (backend → oauth2-proxy:4180)
  - `allow-oauth2-egress-to-google` (oauth2-proxy → any:443 — needed for Google token exchange)
- **Secrets**: chart generates Secrets from `secrets.create.*` values, or references existing via `secrets.existing.*`

## ArgoCD deployment

Application manifest in the infra repo at `infrastructure/argocd/reddwarf/applications/keep.yaml`. Points to `charts/` in this repo via `repoURL: https://github.com/flaccid/google-keep-clone` with inline `valuesObject`.

## CI pipeline

- `lint`: `go vet ./...` + `govulncheck` (continue-on-error)
- `build`: `go build ./...` + store integration tests (Postgres service container)
- `frontend`: `npm ci` + `npm audit --audit-level=moderate` (continue-on-error) + `npm run build`
- `docker`: on `main` only, pushes to Docker Hub `flaccid/google-keep-clone-{backend,frontend}:latest` + `:{sha}`

## Testing

- Store tests (17+5+3) require live PostgreSQL: `DATABASE_URL=postgres://keep:keep@localhost:5432/keep?sslmode=disable go test -count=1 ./store/...`
- API integration tests (9) also require DB
- Frontend: no tests yet (only build check)

## Raw K8s manifests (`k8s/`)

Legacy — `kubectl apply -f k8s/` works but is superseded by the Helm chart. Secret manifests have been removed (chart generates them). Network policies and PDBs are separate files; the chart versions are the canonical source.

## Non-obvious file facts

- `frontend/AGENTS.md` has Next.js 16 rules — read before writing frontend code
- `docs/getting-started.md` and `docs/api-reference.md` supplement the README
- `SECURITY.md` exists locally but is `.gitignore`d (not tracked) — vulnerability notes
- `backend/cmd/keep_server/openapi3.yaml` is COPIED from `gen/http/` by `make gen` — commit it after codegen
- `AGENTS.md` in this repo is the source of truth for how agents should work
