package api

import (
	"context"

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

func (s *MediaService) Download(ctx context.Context, p *media.DownloadPayload) (res []byte, err error) {
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return nil, err
	}
	attachmentID, err := uuid.Parse(p.AttachmentID)
	if err != nil {
		return nil, err
	}
	data, _, err := s.attachmentStore.GetByID(ctx, noteID, attachmentID)
	return data, err
}
