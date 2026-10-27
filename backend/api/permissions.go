package api

import (
	"context"
	"errors"
	"fmt"

	permissions "github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"github.com/flaccid/google-keep-clone/backend/store"
)

type PermissionsService struct {
	permStore *store.PermissionStore
}

func NewPermissionsService(permStore *store.PermissionStore) permissions.Service {
	return &PermissionsService{permStore: permStore}
}

func (s *PermissionsService) BatchCreate(ctx context.Context, p *permissions.BatchCreatePayload) (res []*permissions.Permission, err error) {
	if p == nil || p.BatchCreatePermissionsRequest == nil {
		return nil, permissions.InvalidRequest("permission batch is required")
	}
	owner := store.OwnerFromContext(ctx)
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	res, err = s.permStore.BatchCreate(ctx, owner, noteName, p.BatchCreatePermissionsRequest.Requests)
	return res, permissionError(err)
}

func (s *PermissionsService) BatchDelete(ctx context.Context, p *permissions.BatchDeletePayload) (err error) {
	if p == nil || p.BatchDeletePermissionsRequest == nil {
		return permissions.InvalidRequest("permission batch is required")
	}
	owner := store.OwnerFromContext(ctx)
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	return permissionError(s.permStore.BatchDelete(ctx, owner, noteName, p.BatchDeletePermissionsRequest.Names))
}

func permissionError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, store.ErrInvalidPermission):
		return permissions.InvalidRequest("invalid permission batch")
	case errors.Is(err, store.ErrPermissionDenied):
		return permissions.Forbidden("only the note owner may change permissions")
	case errors.Is(err, store.ErrPermissionNoteNotFound):
		return permissions.NotFound("note not found")
	default:
		return errors.New("permission operation failed")
	}
}
