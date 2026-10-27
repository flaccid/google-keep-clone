package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flaccid/google-keep-clone/backend/gen/permissions"
)

type PermissionStore struct {
	pool *pgxpool.Pool
}

var (
	ErrInvalidPermission      = errors.New("invalid permission request")
	ErrPermissionDenied       = errors.New("only the note owner may change permissions")
	ErrPermissionNoteNotFound = errors.New("note not found")
)

func NewPermissionStore(pool *pgxpool.Pool) *PermissionStore {
	return &PermissionStore{pool: pool}
}

func (s *PermissionStore) BatchCreate(ctx context.Context, owner string, noteName string, requests []*permissions.CreatePermissionRequest) ([]*permissions.Permission, error) {
	noteID, err := parseNoteName(noteName)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid note name", ErrInvalidPermission)
	}
	if err := validatePermissionRequests(requests); err != nil {
		return nil, err
	}
	tx, err := s.ownerTransaction(ctx, owner, noteID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := make([]*permissions.Permission, 0, len(requests))
	for _, req := range requests {
		id := uuid.New()
		role := string(*req.Role)
		email := *req.Email

		err := tx.QueryRow(ctx, `
			INSERT INTO permissions (id, note_id, owner, email, role, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (note_id, email) DO UPDATE SET role = $5
			RETURNING id, email, role
		`, id, noteID, owner, email, role, time.Now().UTC()).Scan(&id, &email, &role)
		if err != nil {
			return nil, fmt.Errorf("insert permission: %w", err)
		}

		r := permissions.Role(role)
		permName := fmt.Sprintf("notes/%s/permissions/%s", noteID, id)
		result = append(result, &permissions.Permission{
			Name:  &permName,
			Role:  &r,
			Email: &email,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit permissions: %w", err)
	}

	return result, nil
}

func (s *PermissionStore) BatchDelete(ctx context.Context, owner string, noteName string, names []string) error {
	noteID, err := parseNoteName(noteName)
	if err != nil {
		return fmt.Errorf("%w: invalid note name", ErrInvalidPermission)
	}

	ids := make([]uuid.UUID, 0, len(names))
	seen := make(map[uuid.UUID]bool)
	for _, permName := range names {
		permID, err := permissionIDForNote(permName, noteID)
		if err != nil {
			return err
		}
		if seen[permID] {
			return fmt.Errorf("%w: duplicate permission", ErrInvalidPermission)
		}
		seen[permID] = true
		ids = append(ids, permID)
	}
	tx, err := s.ownerTransaction(ctx, owner, noteID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, permID := range ids {
		_, err = tx.Exec(ctx, `DELETE FROM permissions WHERE id = $1 AND note_id = $2`, permID, noteID)
		if err != nil {
			return fmt.Errorf("delete permission: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit permissions: %w", err)
	}
	return nil
}

func validatePermissionRequests(requests []*permissions.CreatePermissionRequest) error {
	seen := make(map[string]bool)
	for _, req := range requests {
		if req == nil || req.Role == nil || req.Email == nil || strings.TrimSpace(*req.Email) == "" {
			return fmt.Errorf("%w: role and email are required", ErrInvalidPermission)
		}
		switch string(*req.Role) {
		case "ROLE_UNSPECIFIED", "OWNER", "WRITER", "READER":
		default:
			return fmt.Errorf("%w: unsupported role", ErrInvalidPermission)
		}
		if seen[*req.Email] {
			return fmt.Errorf("%w: duplicate email", ErrInvalidPermission)
		}
		seen[*req.Email] = true
	}
	return nil
}

func permissionIDForNote(name string, noteID uuid.UUID) (uuid.UUID, error) {
	if id, err := uuid.Parse(name); err == nil {
		return id, nil
	}
	parts := strings.Split(name, "/")
	if len(parts) != 4 || parts[0] != "notes" || parts[2] != "permissions" {
		return uuid.Nil, fmt.Errorf("%w: invalid permission name", ErrInvalidPermission)
	}
	parent, err := uuid.Parse(parts[1])
	if err != nil || parent != noteID {
		return uuid.Nil, fmt.Errorf("%w: permission note mismatch", ErrInvalidPermission)
	}
	id, err := uuid.Parse(parts[3])
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid permission ID", ErrInvalidPermission)
	}
	return id, nil
}

func (s *PermissionStore) ownerTransaction(ctx context.Context, owner string, noteID uuid.UUID) (pgx.Tx, error) {
	caller, err := uuid.Parse(owner)
	if err != nil {
		return nil, ErrPermissionDenied
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin permissions: %w", err)
	}
	var actual uuid.UUID
	err = tx.QueryRow(ctx, `SELECT owner FROM notes WHERE id = $1 FOR UPDATE`, noteID).Scan(&actual)
	if err != nil || caller != actual {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPermissionNoteNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("lookup note owner: %w", err)
		}
		return nil, ErrPermissionDenied
	}
	return tx, nil
}
