CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner UUID NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    body_type TEXT CHECK (body_type IN ('text', 'list')),
    body_text TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT 'DEFAULT',
    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    trashed BOOLEAN NOT NULL DEFAULT FALSE,
    trash_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES list_items(id) ON DELETE CASCADE,
    text TEXT NOT NULL DEFAULT '',
    checked BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner UUID NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(owner, display_name)
);

CREATE TABLE note_labels (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, label_id)
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    owner UUID NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('OWNER', 'WRITER')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(note_id, email)
);

CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    mime_type TEXT NOT NULL,
    file_path TEXT NOT NULL,
    byte_size BIGINT NOT NULL DEFAULT 0,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notes_owner ON notes(owner);
CREATE INDEX idx_notes_trashed ON notes(trashed);
CREATE INDEX idx_notes_archived ON notes(archived);
CREATE INDEX idx_notes_pinned ON notes(pinned);
CREATE INDEX idx_notes_created_at ON notes(created_at);
CREATE INDEX idx_notes_updated_at ON notes(updated_at);
CREATE INDEX idx_list_items_note_id ON list_items(note_id);
CREATE INDEX idx_list_items_parent_id ON list_items(parent_id);
CREATE INDEX idx_labels_owner ON labels(owner);
CREATE INDEX idx_permissions_note_id ON permissions(note_id);
CREATE INDEX idx_permissions_owner ON permissions(owner);
CREATE INDEX idx_attachments_note_id ON attachments(note_id);
