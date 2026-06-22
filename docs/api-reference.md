# API Reference

All endpoints are prefixed with `/v1` and proxied via Next.js rewrites (no CORS needed). Interactive docs are available at `/openapi` (Swagger UI) and the raw spec at `/openapi.yaml`.

---

## Notes

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

### Query parameters for `GET /v1/notes`

| Parameter | Type | Description |
|-----------|------|-------------|
| `pageSize` | integer | Notes per page (default: 10) |
| `pageToken` | string | Pagination cursor from previous response |
| `search` | string | Full-text search (ILIKE on `title` + `body_text`) |
| `filter` | string | e.g. `filter=trashed=true`, `filter=archived=true AND trashed=false` |

Valid filter columns: `trashed`, `archived`, `pinned`. Values must be `true` or `false`. Multiple clauses can be joined with ` AND `.

### Note body types

The `body` field is a JSON object with a `type` discriminator:

```json
// Text note
{ "type": "text", "text": "Hello world" }

// Checklist note
{ "type": "list", "listItems": [
  { "text": "apples", "checked": false },
  { "text": "bananas", "checked": true }
] }
```

### Colour values

```
DEFAULT | RED | ORANGE | YELLOW | GREEN | TEAL | BLUE | CERULEAN | PURPLE | PINK | BROWN | GRAY
THEME_SHORE | THEME_BLOOM | THEME_PLUM | THEME_NIGHT | THEME_BAMBOO | THEME_CANDY | THEME_SUNSET | THEME_OCEAN
```

---

## Labels

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/labels` | List all labels |
| `POST` | `/v1/labels` | Create a label |
| `PATCH` | `/v1/labels/{id}` | Rename a label (`{"displayName": "new name"}`) |
| `DELETE` | `/v1/labels/{id}` | Delete a label |

Note labels in API responses are plain display name strings (e.g. `"Groceries"`), not resource names.

---

## Permissions

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/notes/{noteId}/permissions:batchCreate` | Batch create share permissions |
| `POST` | `/v1/notes/{noteId}/permissions:batchDelete` | Batch delete share permissions |

---

## Media (Attachments)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/notes/{noteId}/attachments` | Upload a file (`multipart/form-data`) |
| `GET` | `/v1/notes/{noteId}/attachments/{attachmentId}` | Download an attachment |

Upload accepts the following MIME types:
- `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/svg+xml`
- `application/pdf`
- `text/plain`
- `application/octet-stream`

Unsupported types return `415 Unsupported Media Type`.

Upload uses a standalone `net/http` handler (not Goa-generated) registered on the Goa mux.

---

## OpenAPI

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/openapi` | Swagger UI |
| `GET` | `/openapi.yaml` | Raw OpenAPI 3.0 spec (embedded via `//go:embed`) |

---

## Example: Create a note with curl

```bash
curl -X POST http://localhost:3000/v1/notes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Shopping list",
    "body": {
      "type": "list",
      "listItems": [
        {"text": "milk", "checked": false},
        {"text": "eggs", "checked": true}
      ]
    },
    "color": "TEAL",
    "labels": ["Groceries"]
  }'
```

## Example: Search and filter

```bash
# Search for notes containing "meeting"
curl "http://localhost:3000/v1/notes?search=meeting"

# List archived notes
curl "http://localhost:3000/v1/notes?filter=archived=true"

# Paginated
curl "http://localhost:3000/v1/notes?pageSize=5&pageToken=abc123"
```
