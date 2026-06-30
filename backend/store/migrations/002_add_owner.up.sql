DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notes' AND column_name='owner') THEN
    ALTER TABLE notes ADD COLUMN owner UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
    CREATE INDEX IF NOT EXISTS idx_notes_owner ON notes(owner);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='labels' AND column_name='owner') THEN
    ALTER TABLE labels ADD COLUMN owner UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
    ALTER TABLE labels DROP CONSTRAINT IF EXISTS labels_display_name_key;
    CREATE INDEX IF NOT EXISTS idx_labels_owner ON labels(owner);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='permissions' AND column_name='owner') THEN
    ALTER TABLE permissions ADD COLUMN owner UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
    CREATE INDEX IF NOT EXISTS idx_permissions_owner ON permissions(owner);
  END IF;
END $$;
