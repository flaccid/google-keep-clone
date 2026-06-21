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
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	return s.permStore.BatchCreate(ctx, noteName, p.BatchCreatePermissionsRequest.Requests)
}

func (s *PermissionsService) BatchDelete(ctx context.Context, p *permissions.BatchDeletePayload) (err error) {
	noteName := fmt.Sprintf("notes/%s", p.NoteID)
	return s.permStore.BatchDelete(ctx, noteName, p.BatchDeletePermissionsRequest.Names)
}
