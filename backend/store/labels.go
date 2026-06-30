package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	labels "github.com/flaccid/google-keep-clone/backend/gen/labels"
)

type LabelStore struct {
	pool *pgxpool.Pool
}

func NewLabelStore(pool *pgxpool.Pool) *LabelStore {
	return &LabelStore{pool: pool}
}

func (s *LabelStore) List(ctx context.Context, owner string) ([]*labels.Label, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, display_name FROM labels WHERE owner = $1 ORDER BY display_name`, owner)
	if err != nil {
		return nil, fmt.Errorf("query labels: %w", err)
	}
	defer rows.Close()

	var result []*labels.Label
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		resourceName := fmt.Sprintf("labels/%s", id.String())
		result = append(result, &labels.Label{
			Name:        &resourceName,
			DisplayName: &name,
		})
	}
	return result, nil
}

func (s *LabelStore) Create(ctx context.Context, owner string, displayName string) (*labels.Label, error) {
	id := uuid.New()
	resourceName := fmt.Sprintf("labels/%s", id.String())

	_, err := s.pool.Exec(ctx, `INSERT INTO labels (id, owner, display_name, created_at) VALUES ($1, $2, $3, $4)`,
		id, owner, displayName, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("insert label: %w", err)
	}

	return &labels.Label{
		Name:        &resourceName,
		DisplayName: &displayName,
	}, nil
}

func (s *LabelStore) Update(ctx context.Context, owner string, id string, displayName string) (*labels.Label, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid label id: %w", err)
	}

	_, err = s.pool.Exec(ctx, `UPDATE labels SET display_name = $1 WHERE id = $2 AND owner = $3`, displayName, uid, owner)
	if err != nil {
		return nil, fmt.Errorf("update label: %w", err)
	}

	resourceName := fmt.Sprintf("labels/%s", uid.String())
	return &labels.Label{
		Name:        &resourceName,
		DisplayName: &displayName,
	}, nil
}

func (s *LabelStore) Delete(ctx context.Context, owner string, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid label id: %w", err)
	}

	_, err = s.pool.Exec(ctx, `DELETE FROM labels WHERE id = $1 AND owner = $2`, uid, owner)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	return nil
}
