package api

import (
	"context"
	"fmt"

	media "github.com/flaccid/google-keep-clone/backend/gen/media"
)

type MediaService struct{}

func NewMediaService() media.Service {
	return &MediaService{}
}

func (s *MediaService) Download(ctx context.Context, p *media.DownloadPayload) (res []byte, err error) {
	return nil, fmt.Errorf("not implemented")
}
