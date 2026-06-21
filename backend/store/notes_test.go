package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
)

func TestCreateTextNote(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	title := "Test Note"
	n, err := store.Create(context.Background(), title, strPtr("text"), "Hello World", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, n.Name)
	assert.Contains(t, *n.Name, "notes/")
	assert.Equal(t, title, *n.Title)
	assert.NotNil(t, n.Body)
	assert.NotNil(t, n.Body.Text)
	assert.Equal(t, "Hello World", *n.Body.Text.Text)
	assert.False(t, *n.Pinned)
	assert.False(t, *n.Archived)
	assert.False(t, *n.Trashed)
	assert.NotNil(t, n.CreateTime)
	assert.NotNil(t, n.UpdateTime)
}

func TestCreateNoteWithListItems(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	items := []*notes.ListItem{
		{Text: &notes.TextContent{Text: strPtr("Item 1")}, Checked: boolPtr(false)},
		{Text: &notes.TextContent{Text: strPtr("Item 2")}, Checked: boolPtr(true)},
		{
			Text:    &notes.TextContent{Text: strPtr("Parent")},
			Checked: boolPtr(false),
			ChildListItems: []*notes.ListItem{
				{Text: &notes.TextContent{Text: strPtr("Child 1")}, Checked: boolPtr(false)},
			},
		},
	}

	n, err := store.Create(context.Background(), "List Note", strPtr("list"), "", "DEFAULT", false, false, nil, items)
	require.NoError(t, err)
	require.NotNil(t, n.Body)
	require.NotNil(t, n.Body.List)
	assert.Len(t, n.Body.List.ListItems, 3)
	assert.Equal(t, "Item 1", *n.Body.List.ListItems[0].Text.Text)
	assert.True(t, *n.Body.List.ListItems[1].Checked)
	assert.Len(t, n.Body.List.ListItems[2].ChildListItems, 1)
	assert.Equal(t, "Child 1", *n.Body.List.ListItems[2].ChildListItems[0].Text.Text)
}

func TestCreateNoteWithLabels(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	n, err := store.Create(context.Background(), "Labeled Note", strPtr("text"), "Content", "DEFAULT", false, false, []string{"tag1", "tag2"}, nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"tag1", "tag2"}, n.Labels)
}

func TestGetNoteByID(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Get Me", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	got, err := store.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, *created.Name, *got.Name)
	assert.Equal(t, "Get Me", *got.Title)
}

func TestGetNoteByName(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "By Name", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	got, err := store.GetByName(context.Background(), *created.Name)
	require.NoError(t, err)
	assert.Equal(t, "By Name", *got.Title)
}

func TestListNotes(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	_, err := store.Create(context.Background(), "Note 1", strPtr("text"), "One", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)
	_, err = store.Create(context.Background(), "Note 2", strPtr("text"), "Two", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	res, err := store.List(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, res.Notes, 2)
}

func TestListNotesWithPagination(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	for i := 0; i < 5; i++ {
		_, err := store.Create(context.Background(), "Note", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
		require.NoError(t, err)
	}

	pageSize := 2
	res, err := store.List(context.Background(), &pageSize, nil, nil)
	require.NoError(t, err)
	assert.Len(t, res.Notes, 2)
	require.NotNil(t, res.NextPageToken)

	res2, err := store.List(context.Background(), &pageSize, res.NextPageToken, nil)
	require.NoError(t, err)
	assert.Len(t, res2.Notes, 2)
	require.NotNil(t, res2.NextPageToken)
}

func TestUpdateNoteTitle(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Original", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	newTitle := "Updated"
	updated, err := store.Update(context.Background(), id, &newTitle, nil, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated", *updated.Title)
}

func TestPinUnpinNote(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Pinnable", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	pinned, err := store.SetPinned(context.Background(), id, true)
	require.NoError(t, err)
	assert.True(t, *pinned.Pinned)

	unpinned, err := store.SetPinned(context.Background(), id, false)
	require.NoError(t, err)
	assert.False(t, *unpinned.Pinned)
}

func TestArchiveUnarchive(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Archivable", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	archived, err := store.SetArchived(context.Background(), id, true)
	require.NoError(t, err)
	assert.True(t, *archived.Archived)

	unarchived, err := store.SetArchived(context.Background(), id, false)
	require.NoError(t, err)
	assert.False(t, *unarchived.Archived)
}

func TestTrashRestore(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Trashable", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	trashed, err := store.SetTrashed(context.Background(), id, true)
	require.NoError(t, err)
	assert.True(t, *trashed.Trashed)
	assert.NotNil(t, trashed.TrashTime)

	restored, err := store.SetTrashed(context.Background(), id, false)
	require.NoError(t, err)
	assert.False(t, *restored.Trashed)
	assert.Nil(t, restored.TrashTime)
}

func TestDeleteNote(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	created, err := store.Create(context.Background(), "Deletable", strPtr("text"), "Content", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	id, err := parseNoteName(*created.Name)
	require.NoError(t, err)

	err = store.Delete(context.Background(), id)
	require.NoError(t, err)

	_, err = store.GetByID(context.Background(), id)
	assert.Error(t, err)
}

func TestListTrashedExcluded(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	_, err := store.Create(context.Background(), "Active", strPtr("text"), "Active", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	trashed, err := store.Create(context.Background(), "Trashed", strPtr("text"), "Trashed", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)

	tid, err := parseNoteName(*trashed.Name)
	require.NoError(t, err)
	_, err = store.SetTrashed(context.Background(), tid, true)
	require.NoError(t, err)

	res, err := store.List(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Len(t, res.Notes, 1)
	assert.Equal(t, "Active", *res.Notes[0].Title)
}

func TestNoteNotFound(t *testing.T) {
	pool := newTestPool(t)
	store := NewNoteStore(pool)

	_, err := store.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
