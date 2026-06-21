package api

import (
	"context"

	labels "github.com/flaccid/google-keep-clone/backend/gen/labels"
	"github.com/flaccid/google-keep-clone/backend/store"
)

type LabelsService struct {
	labelStore *store.LabelStore
}

func NewLabelsService(labelStore *store.LabelStore) labels.Service {
	return &LabelsService{labelStore: labelStore}
}

func (s *LabelsService) List(ctx context.Context) (res []*labels.Label, err error) {
	return s.labelStore.List(ctx)
}

func (s *LabelsService) Create(ctx context.Context, p *labels.CreatePayload) (res *labels.Label, err error) {
	return s.labelStore.Create(ctx, p.DisplayName)
}

func (s *LabelsService) Delete(ctx context.Context, p *labels.DeletePayload) (err error) {
	return s.labelStore.Delete(ctx, p.ID)
}
