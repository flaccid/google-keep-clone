package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flaccid/google-keep-clone/backend/gen/permissions"
)

type PermissionStore struct {
	pool *pgxpool.Pool
}

func NewPermissionStore(pool *pgxpool.Pool) *PermissionStore {
	return &PermissionStore{pool: pool}
}

func (s *PermissionStore) BatchCreate(ctx context.Context, noteName string, requests []*permissions.CreatePermissionRequest) ([]*permissions.Permission, error) {
	noteID, err := parseNoteName(noteName)
	if err != nil {
		return nil, fmt.Errorf("invalid note name: %w", err)
	}

	var result []*permissions.Permission
	for _, req := range requests {
		id := uuid.New()
		permName := fmt.Sprintf("%s/permissions/%s", noteName, id.String())
		role := string(*req.Role)

		_, err := s.pool.Exec(ctx, `
			INSERT INTO permissions (id, note_id, email, role, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (note_id, email) DO UPDATE SET role = $4
		`, id, noteID, *req.Email, role, time.Now().UTC())
		if err != nil {
			return nil, fmt.Errorf("insert permission: %w", err)
		}

		r := permissions.Role(role)
		email := *req.Email
		result = append(result, &permissions.Permission{
			Name:  &permName,
			Role:  &r,
			Email: &email,
		})
	}

	return result, nil
}

func (s *PermissionStore) BatchDelete(ctx context.Context, noteName string, names []string) error {
	noteID, err := parseNoteName(noteName)
	if err != nil {
		return fmt.Errorf("invalid note name: %w", err)
	}

	for _, permName := range names {
		permID, err := parsePermissionName(permName)
		if err != nil {
			return fmt.Errorf("invalid permission name: %w", err)
		}
		_, err = s.pool.Exec(ctx, `DELETE FROM permissions WHERE id = $1 AND note_id = $2`, permID, noteID)
		if err != nil {
			return fmt.Errorf("delete permission: %w", err)
		}
	}

	return nil
}

func parsePermissionName(name string) (uuid.UUID, error) {
	parts := []rune(name)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			return uuid.Parse(string(parts[i+1:]))
		}
	}
	return uuid.Parse(name)
}
