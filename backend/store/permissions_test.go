package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flaccid/google-keep-clone/backend/gen/permissions"
)

func TestBatchCreatePermissions(t *testing.T) {
	pool := newTestPool(t)
	noteStore := NewNoteStore(pool)
	permStore := NewPermissionStore(pool)

	n, err := noteStore.Create(context.Background(), "Shared Note", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)
	noteName := *n.Name

	roleWriter := permissions.Role("WRITER")
	roleReader := permissions.Role("OWNER")
	requests := []*permissions.CreatePermissionRequest{
		{Role: &roleWriter, Email: strPtr("alice@example.com")},
		{Role: &roleReader, Email: strPtr("bob@example.com")},
	}

	results, err := permStore.BatchCreate(context.Background(), noteName, requests)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, *results[0].Name, noteName+"/permissions/")
	assert.Equal(t, "alice@example.com", *results[0].Email)
	assert.Equal(t, permissions.Role("WRITER"), *results[0].Role)
}

func TestBatchCreateDuplicateEmail(t *testing.T) {
	pool := newTestPool(t)
	noteStore := NewNoteStore(pool)
	permStore := NewPermissionStore(pool)

	n, err := noteStore.Create(context.Background(), "Shared Note", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	roleReader := permissions.Role("OWNER")
	roleWriter := permissions.Role("WRITER")

	// Create first
	_, err = permStore.BatchCreate(context.Background(), *n.Name, []*permissions.CreatePermissionRequest{
		{Role: &roleReader, Email: strPtr("alice@example.com")},
	})
	require.NoError(t, err)

	// Upsert with different role
	results, err := permStore.BatchCreate(context.Background(), *n.Name, []*permissions.CreatePermissionRequest{
		{Role: &roleWriter, Email: strPtr("alice@example.com")},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, permissions.Role("WRITER"), *results[0].Role)
}

func TestBatchDeletePermissions(t *testing.T) {
	pool := newTestPool(t)
	noteStore := NewNoteStore(pool)
	permStore := NewPermissionStore(pool)

	n, err := noteStore.Create(context.Background(), "Shared Note", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	roleReader := permissions.Role("OWNER")
	results, err := permStore.BatchCreate(context.Background(), *n.Name, []*permissions.CreatePermissionRequest{
		{Role: &roleReader, Email: strPtr("alice@example.com")},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)

	permName := *results[0].Name
	err = permStore.BatchDelete(context.Background(), *n.Name, []string{permName})
	require.NoError(t, err)

	// Permission should no longer exist in DB
	err = permStore.BatchDelete(context.Background(), *n.Name, []string{permName})
	require.NoError(t, err)
}

func TestBatchDeleteInvalidNote(t *testing.T) {
	pool := newTestPool(t)
	permStore := NewPermissionStore(pool)

	err := permStore.BatchDelete(context.Background(), "notes/"+fmt.Sprintf("%d", 0), nil)
	require.Error(t, err)
}
