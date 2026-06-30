DROP INDEX IF EXISTS idx_permissions_owner;
ALTER TABLE permissions DROP COLUMN owner;

ALTER TABLE labels DROP COLUMN owner;
ALTER TABLE labels ADD UNIQUE (display_name);

DROP INDEX IF EXISTS idx_notes_owner;
ALTER TABLE notes DROP COLUMN owner;
