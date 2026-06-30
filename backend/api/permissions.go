package api

import (
	"context"
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
	owner := store.OwnerFromContext(ctx)
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	return s.permStore.BatchCreate(ctx, owner, noteName, p.BatchCreatePermissionsRequest.Requests)
}

func (s *PermissionsService) BatchDelete(ctx context.Context, p *permissions.BatchDeletePayload) (err error) {
	owner := store.OwnerFromContext(ctx)
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	return s.permStore.BatchDelete(ctx, owner, noteName, p.BatchDeletePermissionsRequest.Names)
}
