package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	media "github.com/flaccid/google-keep-clone/backend/gen/media"
	"github.com/flaccid/google-keep-clone/backend/store"
)

type MediaService struct {
	attachmentStore *store.AttachmentStore
}

func NewMediaService(attachmentStore *store.AttachmentStore) media.Service {
	return &MediaService{attachmentStore: attachmentStore}
}

func (s *MediaService) Upload(ctx context.Context, p *media.UploadPayload) (res *media.Attachment, err error) {
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return nil, fmt.Errorf("invalid note ID: %w", err)
	}

	contentType := p.ContentType
	if contentType == "" || len(contentType) > 256 {
		contentType = "application/octet-stream"
	}
	allowed := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/svg+xml": true, "application/pdf": true,
		"text/plain": true, "application/octet-stream": true,
	}
	if !allowed[contentType] {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	storeAtt, err := s.attachmentStore.Upload(ctx, noteID, contentType, p.Data)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	return &media.Attachment{
		Name:     storeAtt.Name,
		MimeType: storeAtt.MimeType,
	}, nil
}

func (s *MediaService) Download(ctx context.Context, p *media.DownloadPayload) (res []byte, err error) {
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return nil, err
	}
	attachmentID, err := uuid.Parse(p.AttachmentID)
	if err != nil {
		return nil, err
	}
	data, mime, err := s.attachmentStore.GetByID(ctx, noteID, attachmentID)
	if err != nil {
		return nil, err
	}
	if p.MimeType != nil && *p.MimeType != "" && *p.MimeType != mime {
		return nil, fmt.Errorf("attachment not available in requested MIME type: %s", *p.MimeType)
	}
	return data, err
}
