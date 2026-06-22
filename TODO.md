# TODO

> Project improvements and nice-to-haves. Not blocking, but worth doing.

## Security & CI

- [ ] Enable **Dependabot** for Go modules (`backend/`) and npm (`frontend/`) — automated PRs for dependency updates
- [ ] Add **`govulncheck`** to the backend CI pipeline (`go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...`)
- [ ] Add **`npm audit`** to the frontend CI pipeline
- [ ] Move CI database credentials from plaintext in `.github/workflows/ci.yml` to **GitHub Actions Secrets**

## Testing

- [ ] Add **frontend tests** (unit + component tests)
- [ ] Add **end-to-end tests** for critical flows (create note, archive, trash, search, labels)
- [ ] Run backend API integration tests in CI (currently only store tests run)
- [ ] Add API-level **integration tests** for the Go HTTP handlers

## Features

- [ ] Show label chips on `NoteCard` for at-a-glance label visibility
- [ ] Add dedicated endpoint for adding/removing a single label from a note
- [ ] Add **note reminders** (the reminders page is a placeholder)
- [ ] Implement **note archiving from the card** (pin/archive buttons on hover)
- [ ] Add **undo snackbar** after trashing/deleting a note

## UX Polish

- [ ] Add **empty state illustrations** for archive, trash, and labels pages
- [ ] Add **keyboard shortcuts** (e.g. `Ctrl+N` new note, `Escape` close modal)
- [ ] Add **loading skeletons** for note grid
- [ ] Add **toast notifications** for save/delete/restore actions

## Infrastructure

- [ ] Migrate from plain K8s manifests to a **Helm chart** for easier deployment
- [ ] Add **Sealed Secrets** or **External Secrets Operator** for managing secrets in-cluster
- [ ] Add **Horizontal Pod Autoscaler** for the frontend
- [ ] Add **network policies** for the K8s namespace
- [ ] Add **PodDisruptionBudget** for production deployments

## Documentation

- [ ] Add **architecture decision records** (ADRs) in `docs/adr/`
- [ ] Add **API usage examples** (curl commands) in the README
- [ ] Add **screenshots** to the README showing the UI
