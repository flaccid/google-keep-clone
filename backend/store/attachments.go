package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
)

type AttachmentStore struct {
	pool    *pgxpool.Pool
	storeDir string
}

func NewAttachmentStore(pool *pgxpool.Pool) *AttachmentStore {
	dir := os.Getenv("ATTACHMENT_STORE_DIR")
	if dir == "" {
		dir = "./attachments"
	}
	os.MkdirAll(dir, 0755)
	return &AttachmentStore{pool: pool, storeDir: dir}
}

func (s *AttachmentStore) Upload(ctx context.Context, noteID uuid.UUID, contentType string, data []byte) (*notes.Attachment, error) {
	id := uuid.New()
	filename := id.String()
	filePath := filepath.Join(s.storeDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("write attachment file: %w", err)
	}

	noteName := fmt.Sprintf("notes/%s", noteID.String())
	attachmentName := fmt.Sprintf("%s/attachments/%s", noteName, id.String())

	_, err := s.pool.Exec(ctx, `
		INSERT INTO attachments (id, note_id, mime_type, file_path, byte_size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, noteID, contentType, filePath, int64(len(data)), time.Now().UTC())
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("insert attachment: %w", err)
	}

	return &notes.Attachment{
		Name:     &attachmentName,
		MimeType: []string{contentType},
	}, nil
}

type AttachmentMeta struct {
	Name     string
	MimeType []string
}

func (s *AttachmentStore) GetMetaByID(ctx context.Context, noteID, attachmentID uuid.UUID) (*AttachmentMeta, error) {
	var mimeType string
	err := s.pool.QueryRow(ctx, `
		SELECT mime_type FROM attachments WHERE id = $1 AND note_id = $2
	`, attachmentID, noteID).Scan(&mimeType)
	if err != nil {
		return nil, fmt.Errorf("query attachment: %w", err)
	}
	name := fmt.Sprintf("notes/%s/attachments/%s", noteID.String(), attachmentID.String())
	return &AttachmentMeta{
		Name:     name,
		MimeType: []string{mimeType},
	}, nil
}

func (s *AttachmentStore) GetByID(ctx context.Context, noteID, attachmentID uuid.UUID) ([]byte, string, error) {
	var filePath string
	var mimeType string
	err := s.pool.QueryRow(ctx, `
		SELECT file_path, mime_type FROM attachments WHERE id = $1 AND note_id = $2
	`, attachmentID, noteID).Scan(&filePath, &mimeType)
	if err != nil {
		return nil, "", fmt.Errorf("query attachment: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("read attachment file: %w", err)
	}

	return data, mimeType, nil
}

func (s *AttachmentStore) ListByNote(ctx context.Context, noteID uuid.UUID) ([]*notes.Attachment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, mime_type FROM attachments WHERE note_id = $1 ORDER BY created_at
	`, noteID)
	if err != nil {
		return nil, fmt.Errorf("query attachments: %w", err)
	}
	defer rows.Close()

	var result []*notes.Attachment
	for rows.Next() {
		var id uuid.UUID
		var mime string
		if err := rows.Scan(&id, &mime); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		name := fmt.Sprintf("notes/%s/attachments/%s", noteID.String(), id.String())
		result = append(result, &notes.Attachment{
			Name:     &name,
			MimeType: []string{mime},
		})
	}
	return result, nil
}
