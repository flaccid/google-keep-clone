package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLabel(t *testing.T) {
	pool := newTestPool(t)
	store := NewLabelStore(pool)

	l, err := store.Create(context.Background(), "Test Label")
	require.NoError(t, err)
	require.NotNil(t, l.Name)
	assert.Contains(t, *l.Name, "labels/")
	assert.Equal(t, "Test Label", *l.DisplayName)
}

func TestListLabels(t *testing.T) {
	pool := newTestPool(t)
	store := NewLabelStore(pool)

	_, err := store.Create(context.Background(), "B")
	require.NoError(t, err)
	_, err = store.Create(context.Background(), "A")
	require.NoError(t, err)

	labels, err := store.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "A", *labels[0].DisplayName)
	assert.Equal(t, "B", *labels[1].DisplayName)
}

func TestDeleteLabel(t *testing.T) {
	pool := newTestPool(t)
	store := NewLabelStore(pool)

	l, err := store.Create(context.Background(), "To Delete")
	require.NoError(t, err)

	err = store.Delete(context.Background(), "does-not-exist")
	require.Error(t, err)

	id := (*l.Name)[7:]
	err = store.Delete(context.Background(), id)
	require.NoError(t, err)

	labels, err := store.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, labels, 0)
}
