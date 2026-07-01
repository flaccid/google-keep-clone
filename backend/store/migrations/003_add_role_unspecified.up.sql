DO $$
BEGIN
    ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_role_check;
    ALTER TABLE permissions ADD CONSTRAINT permissions_role_check CHECK (role IN ('ROLE_UNSPECIFIED', 'OWNER', 'WRITER'));
END $$;
