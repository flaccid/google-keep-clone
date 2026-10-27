# Permission mutation contract

Permission batches operate on one note and are authorized by its **persisted
owner UUID**, not a collaborator's role label, email or stored grantor field.
An OWNER-labelled collaborator does not acquire note ownership. Embedders must
apply any additional resource-authority policy before invoking these services;
this module does not implement linked-resource synchronization or an outbox.

## Atomic operations

- A create/update or delete batch uses one database transaction and locks the
  note row. Validation, authorization and statement failures commit no prefix of
  the batch. Results are returned only after successful commit acknowledgement;
  transport failure during acknowledgement can still require a state read/retry.
- Upsert identity is `(note_id, email)`. Returned resource names contain the
  actual persisted UUID, including when an existing row changes role.
- Emails are matched exactly, without new normalization or account merging.
  Blank emails, missing role/email fields, unknown roles and duplicate exact
  emails in a batch are rejected. Existing roles remain supported; READER is added.
- Delete accepts `notes/{note UUID}/permissions/{permission UUID}`. Names must
  belong to the requested note. Bare permission UUIDs remain supported for
  compatibility; arbitrary path prefixes do not. Duplicate IDs (even with mixed
  bare/full spellings) are rejected before any write.
- Once note ownership is verified, missing permission IDs are idempotent success.
  Existing grants cannot escape revocation because their historical grantor field
  differs from the actual note owner. No automatic grantor backfill is performed.
- Empty batches are noops **only after** verifying note ownership. Invalid batches
  return HTTP 400, non-owner mutations 403, missing notes 404. Unexpected database
  failures return a sanitized 500 and no partial result collection.

READER grants allow the existing note read/list paths, not content changes,
lifecycle writes, permanent deletion or permission mutations. WRITER content
editing does not confer sharing-management authority. This change does not claim
to implement authentication or per-note authorization for every separate transport
or attachment integration; embedders must retain their existing guards.

## Schema and release compatibility

Migration 005 expands the permission CHECK to admit READER without rewriting
existing grants. Downgrade refuses to remove reader support while reader rows
exist: it does not promote, drop or merge them. Review backup/rollback and
application compatibility before applying schema migrations to any deployed DB.
An attempted blocked downgrade may leave the migration tool's version marked
dirty and requires reviewed recovery; do not force versions automatically.
The server's normal migration-on-startup behavior remains unchanged.

Consumers with independently owned schemas must arrange their own reviewed CHECK
update; updating the Go dependency alone is insufficient. The added enum member,
typed error responses, stricter batch validation, owner-only store enforcement
and atomicity are compatibility changes to review before a versioned release.

## Generation and tests

Run `make gen` from `backend/`; never hand-edit generated files. The generation
step reapplies source-controlled overlays in `tools/codegen/templates` to preserve
the public OpenAPI export and existing raw-media response encoding. Compatibility
tests detect loss of those contracts.

Against a disposable PostgreSQL database, run the complete backend suite with
`go test -race -p 1 -count=1 ./...`. Packages are serialized because existing
integration fixtures truncate shared test tables. Tests cover stable IDs,
reader downgrade/read-only behavior, owner/peer boundaries, different-grantor
revocation, validation, SQL-failure rollback, concurrent upserts, migration
upgrade/downgrade and generated HTTP status/encoding contracts.
