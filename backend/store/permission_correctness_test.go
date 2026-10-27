package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func permissionFixture(t *testing.T) (*pgxpool.Pool, *PermissionStore, *NoteStore, string) {
	t.Helper()
	pool := newTestPool(t)
	notes := NewNoteStore(pool, nil)
	note, err := notes.Create(context.Background(), testOwner, "permission fixture", strPtr("text"), "synthetic text", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)
	return pool, NewPermissionStore(pool), notes, *note.Name
}

func permissionRequest(email, role string) *permissions.CreatePermissionRequest {
	r := permissions.Role(role)
	return &permissions.CreatePermissionRequest{Role: &r, Email: &email}
}

func permissionRows(t *testing.T, pool *pgxpool.Pool, note string) int {
	t.Helper()
	id, err := parseNoteName(note)
	require.NoError(t, err)
	var count int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM permissions WHERE note_id=$1`, id).Scan(&count))
	return count
}

func TestPermissionValidationBeforeMutation(t *testing.T) {
	s := NewPermissionStore(nil)
	note := "notes/" + uuid.NewString()
	for _, requests := range [][]*permissions.CreatePermissionRequest{
		{nil}, {{}}, {{Role: ptr(permissions.Role("WRITER"))}},
		{permissionRequest(" ", "WRITER")}, {permissionRequest("peer@example.invalid", "INVALID")},
		{permissionRequest("peer@example.invalid", "WRITER"), permissionRequest("peer@example.invalid", "READER")},
	} {
		_, err := s.BatchCreate(context.Background(), testOwner, note, requests)
		require.ErrorIs(t, err, ErrInvalidPermission)
	}
	id := uuid.NewString()
	for _, names := range [][]string{
		{"invalid"}, {"arbitrary/prefix/" + id}, {"notes/" + uuid.NewString() + "/permissions/" + id},
		{note + "/permissions/" + id, id},
	} {
		require.ErrorIs(t, s.BatchDelete(context.Background(), testOwner, note, names), ErrInvalidPermission)
	}
	_, err := s.BatchCreate(context.Background(), "invalid-caller", note, nil)
	require.ErrorIs(t, err, ErrPermissionDenied)
}

func TestPermissionReaderAndPersistedUpsertIdentity(t *testing.T) {
	pool, perms, notes, note := permissionFixture(t)
	ctx := context.Background()
	first, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("reader@example.invalid", "WRITER")})
	require.NoError(t, err)
	second, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("reader@example.invalid", "READER")})
	require.NoError(t, err)
	require.Equal(t, *first[0].Name, *second[0].Name)
	require.Equal(t, permissions.Role("READER"), *second[0].Role)
	id, err := parseNoteName(note)
	require.NoError(t, err)
	var persisted string
	require.NoError(t, pool.QueryRow(ctx, `SELECT id::text FROM permissions WHERE note_id=$1 AND email='reader@example.invalid'`, id).Scan(&persisted))
	require.Equal(t, note+"/permissions/"+persisted, *second[0].Name)
	reader := WithEmail(ctx, "reader@example.invalid")
	peer := uuid.NewString()
	_, err = notes.GetByName(reader, peer, note)
	require.NoError(t, err)
	listed, err := notes.List(reader, peer, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, listed.Notes, 1)
	_, err = notes.Update(reader, peer, id, strPtr("forbidden reader write"), nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	_, err = notes.SetPinned(reader, peer, id, true)
	require.Error(t, err)
	_, err = notes.SetArchived(reader, peer, id, true)
	require.Error(t, err)
	_, err = notes.SetTrashed(reader, peer, id, true)
	require.Error(t, err)
	require.Error(t, notes.Delete(reader, peer, id))
	require.NoError(t, perms.BatchDelete(ctx, testOwner, note, []string{*second[0].Name}))
	require.Equal(t, 0, permissionRows(t, pool, note))
	_, err = notes.GetByName(reader, peer, note)
	require.Error(t, err)
	require.NoError(t, perms.BatchDelete(ctx, testOwner, note, []string{*second[0].Name}))
}

func TestPermissionOnlyActualOwnerMutatesIncludingEmptyBatches(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	ctx := context.Background()
	for _, role := range []string{"WRITER", "READER", "OWNER"} {
		_, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest(strings.ToLower(role)+"@example.invalid", role)})
		require.NoError(t, err)
	}
	for _, role := range []string{"writer", "reader", "owner", "stranger"} {
		caller := WithEmail(ctx, role+"@example.invalid")
		for _, requests := range [][]*permissions.CreatePermissionRequest{nil, {permissionRequest("new@example.invalid", "WRITER")}} {
			_, err := perms.BatchCreate(caller, uuid.NewString(), note, requests)
			require.ErrorIs(t, err, ErrPermissionDenied)
		}
		require.ErrorIs(t, perms.BatchDelete(caller, uuid.NewString(), note, nil), ErrPermissionDenied)
	}
	require.Equal(t, 3, permissionRows(t, pool, note))
	_, err := perms.BatchCreate(ctx, testOwner, "notes/"+uuid.NewString(), nil)
	require.ErrorIs(t, err, ErrPermissionNoteNotFound)
	require.ErrorIs(t, perms.BatchDelete(ctx, testOwner, "notes/"+uuid.NewString(), nil), ErrPermissionNoteNotFound)
}

func TestPermissionActualOwnerRevokesDifferentGrantor(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	ctx := context.Background()
	noteID, err := parseNoteName(note)
	require.NoError(t, err)
	id, grantor := uuid.New(), uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO permissions (id,note_id,owner,email,role) VALUES ($1,$2,$3,'peer@example.invalid','WRITER')`, id, noteID, grantor)
	require.NoError(t, err)
	updated, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("peer@example.invalid", "READER")})
	require.NoError(t, err)
	require.Equal(t, note+"/permissions/"+id.String(), *updated[0].Name)
	require.ErrorIs(t, perms.BatchDelete(ctx, grantor.String(), note, []string{*updated[0].Name}), ErrPermissionDenied)
	require.NoError(t, perms.BatchDelete(ctx, testOwner, note, []string{id.String()}))
	require.Equal(t, 0, permissionRows(t, pool, note))
}

func TestPermissionBatchesRollbackValidationAndSQLFailures(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	ctx := context.Background()
	requests := []*permissions.CreatePermissionRequest{permissionRequest("first@example.invalid", "WRITER"), permissionRequest("second@example.invalid", "WRITER")}
	_, err := perms.BatchCreate(ctx, testOwner, note, append(requests, nil))
	require.ErrorIs(t, err, ErrInvalidPermission)
	require.Zero(t, permissionRows(t, pool, note))
	id, err := parseNoteName(note)
	require.NoError(t, err)
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	name := "permission_fixture_" + suffix
	_, err = pool.Exec(ctx, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN IF COALESCE(NEW.note_id, OLD.note_id) = '%s'::uuid AND COALESCE(NEW.email, OLD.email) = 'second@example.invalid' THEN
		RAISE EXCEPTION 'synthetic batch failure'; END IF; RETURN COALESCE(NEW, OLD); END $$;
		CREATE TRIGGER %s BEFORE INSERT OR UPDATE OR DELETE ON permissions FOR EACH ROW EXECUTE FUNCTION %s()`, name, id, name, name))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := pool.Exec(ctx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON permissions; DROP FUNCTION IF EXISTS %s()`, name, name))
		require.NoError(t, err)
	})
	result, err := perms.BatchCreate(ctx, testOwner, note, requests)
	require.Error(t, err)
	require.Nil(t, result)
	require.Zero(t, permissionRows(t, pool, note))
	_, err = pool.Exec(ctx, fmt.Sprintf(`ALTER TABLE permissions DISABLE TRIGGER %s`, name))
	require.NoError(t, err)
	created, err := perms.BatchCreate(ctx, testOwner, note, requests)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, fmt.Sprintf(`ALTER TABLE permissions ENABLE TRIGGER %s`, name))
	require.NoError(t, err)
	names := []string{*created[0].Name, *created[1].Name}
	require.Error(t, perms.BatchDelete(ctx, testOwner, note, names))
	require.Equal(t, 2, permissionRows(t, pool, note))
	require.ErrorIs(t, perms.BatchDelete(ctx, testOwner, note, []string{names[0], "invalid"}), ErrInvalidPermission)
	require.Equal(t, 2, permissionRows(t, pool, note))
	result, err = perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("first@example.invalid", "READER"), permissionRequest("second@example.invalid", "READER")})
	require.Error(t, err)
	require.Nil(t, result)
	var role string
	require.NoError(t, pool.QueryRow(ctx, `SELECT role FROM permissions WHERE note_id=$1 AND email='first@example.invalid'`, id).Scan(&role))
	require.Equal(t, "WRITER", role)
}

func TestPermissionConcurrentUpsertsReturnOnePersistedID(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	type outcome struct {
		permissions []*permissions.Permission
		err         error
	}
	results := make(chan outcome, 16)
	var group sync.WaitGroup
	for range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			res, err := perms.BatchCreate(context.Background(), testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("peer@example.invalid", "READER")})
			results <- outcome{res, err}
		}()
	}
	group.Wait()
	close(results)
	var name string
	for result := range results {
		require.NoError(t, result.err)
		require.Len(t, result.permissions, 1)
		if name == "" {
			name = *result.permissions[0].Name
		}
		require.Equal(t, name, *result.permissions[0].Name)
	}
	require.Equal(t, 1, permissionRows(t, pool, note))
}

func TestPermissionReaderMigrationRollbackSafety(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	ctx := context.Background()
	up, err := migrationsFS.ReadFile("migrations/005_add_reader_role.up.sql")
	require.NoError(t, err)
	down, err := migrationsFS.ReadFile("migrations/005_add_reader_role.down.sql")
	require.NoError(t, err)
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()
	t.Cleanup(func() { _, err := pool.Exec(ctx, string(up)); require.NoError(t, err) })
	grants, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("peer@example.invalid", "READER")})
	require.NoError(t, err)
	_, err = conn.Exec(ctx, string(down))
	require.Error(t, err)
	_, err = conn.Exec(ctx, "ROLLBACK")
	require.NoError(t, err)
	require.NoError(t, perms.BatchDelete(ctx, testOwner, note, []string{*grants[0].Name}))
	_, err = conn.Exec(ctx, string(down))
	require.NoError(t, err)
	_, err = perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("peer@example.invalid", "READER")})
	require.Error(t, err)
	_, err = conn.Exec(ctx, string(up))
	require.NoError(t, err)
	_, err = perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("peer@example.invalid", "READER")})
	require.NoError(t, err)
}

func TestPermissionReaderUpgradePreservesExistingGrants(t *testing.T) {
	pool, perms, _, note := permissionFixture(t)
	ctx := context.Background()
	grants, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("writer@example.invalid", "WRITER")})
	require.NoError(t, err)
	source, err := iofs.New(migrationsFS, "migrations")
	require.NoError(t, err)
	m, err := migrate.NewWithSourceInstance("iofs", source, getTestURL(t))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, RunMigrations(getTestURL(t)))
		sourceErr, databaseErr := m.Close()
		require.NoError(t, sourceErr)
		require.NoError(t, databaseErr)
	})
	require.NoError(t, m.Steps(-1))
	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.EqualValues(t, 4, version)
	require.False(t, dirty)
	require.NoError(t, RunMigrations(getTestURL(t)))
	version, dirty, err = m.Version()
	require.NoError(t, err)
	require.EqualValues(t, 5, version)
	require.False(t, dirty)
	updated, err := perms.BatchCreate(ctx, testOwner, note, []*permissions.CreatePermissionRequest{permissionRequest("writer@example.invalid", "READER")})
	require.NoError(t, err)
	require.Equal(t, *grants[0].Name, *updated[0].Name)
	require.Equal(t, 1, permissionRows(t, pool, note))
}
