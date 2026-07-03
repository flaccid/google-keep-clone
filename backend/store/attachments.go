package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
)

// AttachmentStore is the interface for storing and retrieving attachment data.
type AttachmentStore interface {
	Upload(ctx context.Context, noteID uuid.UUID, contentType string, data []byte) (*notes.Attachment, error)
	GetMetaByID(ctx context.Context, noteID, attachmentID uuid.UUID) (*AttachmentMeta, error)
	GetByID(ctx context.Context, noteID, attachmentID uuid.UUID) ([]byte, string, error)
	ListByNote(ctx context.Context, noteID uuid.UUID) ([]*notes.Attachment, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
}

type AttachmentMeta struct {
	Name     string
	MimeType []string
}

// NewAttachmentStore creates the appropriate AttachmentStore based on environment variables.
// If S3_ENDPOINT is set, an S3-compatible store is created; otherwise the local filesystem store is used.
func NewAttachmentStore(pool *pgxpool.Pool) AttachmentStore {
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		return newS3Store(pool)
	}
	return newFilesystemStore(pool)
}

// --- Filesystem implementation ---

type filesystemStore struct {
	pool     *pgxpool.Pool
	storeDir string
}

func newFilesystemStore(pool *pgxpool.Pool) *filesystemStore {
	dir := os.Getenv("ATTACHMENT_STORE_DIR")
	if dir == "" {
		dir = "./attachments"
	}
	os.MkdirAll(dir, 0755)
	return &filesystemStore{pool: pool, storeDir: dir}
}

func (s *filesystemStore) Upload(ctx context.Context, noteID uuid.UUID, contentType string, data []byte) (*notes.Attachment, error) {
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

func (s *filesystemStore) GetMetaByID(ctx context.Context, noteID, attachmentID uuid.UUID) (*AttachmentMeta, error) {
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

func (s *filesystemStore) GetByID(ctx context.Context, noteID, attachmentID uuid.UUID) ([]byte, string, error) {
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

func (s *filesystemStore) ListByNote(ctx context.Context, noteID uuid.UUID) ([]*notes.Attachment, error) {
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

func (s *filesystemStore) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	rows, err := s.pool.Query(ctx, `SELECT file_path FROM attachments WHERE note_id = $1`, noteID)
	if err != nil {
		return fmt.Errorf("query attachments for deletion: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return fmt.Errorf("scan attachment file_path: %w", err)
		}
		os.Remove(filePath)
	}
	return nil
}

// --- S3 implementation ---

type s3Store struct {
	pool   *pgxpool.Pool
	client *minio.Client
	bucket string
}

func newS3Store(pool *pgxpool.Pool) *s3Store {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	useSSL := os.Getenv("S3_USE_SSL") != "false"

	if bucket == "" {
		bucket = "attachments"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create S3 client: %v", err))
	}

	// Ensure bucket exists
	ctx := context.Background()
	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		exists, errExists := client.BucketExists(ctx, bucket)
		if errExists != nil {
			// Both MakeBucket and BucketExists failed — likely a permission
			// restriction (e.g. Backblaze B2 application key lacks
			// createBucket / listBuckets). The bucket already exists and
			// PutObject/GetObject still work, so proceed with a warning.
			fmt.Fprintf(os.Stderr, "WARN: unable to verify S3 bucket %q: %v (MakeBucket: %v); assuming it exists\n", bucket, errExists, err)
		} else if !exists {
			panic(fmt.Sprintf("S3 bucket %s does not exist", bucket))
		}
	}

	return &s3Store{pool: pool, client: client, bucket: bucket}
}

func (s *s3Store) Upload(ctx context.Context, noteID uuid.UUID, contentType string, data []byte) (*notes.Attachment, error) {
	id := uuid.New()
	objectKey := fmt.Sprintf("%s/%s", noteID.String(), id.String())

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 upload: %w", err)
	}

	noteName := fmt.Sprintf("notes/%s", noteID.String())
	attachmentName := fmt.Sprintf("%s/attachments/%s", noteName, id.String())

	_, err = s.pool.Exec(ctx, `
		INSERT INTO attachments (id, note_id, mime_type, file_path, byte_size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, noteID, contentType, objectKey, int64(len(data)), time.Now().UTC())
	if err != nil {
		// Clean up S3 object on DB insert failure
		s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
		return nil, fmt.Errorf("insert attachment: %w", err)
	}

	return &notes.Attachment{
		Name:     &attachmentName,
		MimeType: []string{contentType},
	}, nil
}

func (s *s3Store) GetMetaByID(ctx context.Context, noteID, attachmentID uuid.UUID) (*AttachmentMeta, error) {
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

func (s *s3Store) GetByID(ctx context.Context, noteID, attachmentID uuid.UUID) ([]byte, string, error) {
	var filePath string
	var mimeType string
	err := s.pool.QueryRow(ctx, `
		SELECT file_path, mime_type FROM attachments WHERE id = $1 AND note_id = $2
	`, attachmentID, noteID).Scan(&filePath, &mimeType)
	if err != nil {
		return nil, "", fmt.Errorf("query attachment: %w", err)
	}

	obj, err := s.client.GetObject(ctx, s.bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("s3 get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("s3 read object: %w", err)
	}

	return data, mimeType, nil
}

func (s *s3Store) ListByNote(ctx context.Context, noteID uuid.UUID) ([]*notes.Attachment, error) {
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

func (s *s3Store) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	rows, err := s.pool.Query(ctx, `SELECT file_path FROM attachments WHERE note_id = $1`, noteID)
	if err != nil {
		return fmt.Errorf("query attachments for deletion: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return fmt.Errorf("scan attachment file_path: %w", err)
		}
		if err := s.client.RemoveObject(ctx, s.bucket, filePath, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove S3 object: %w", err)
		}
	}
	return nil
}
