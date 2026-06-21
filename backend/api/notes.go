package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
	"github.com/flaccid/google-keep-clone/backend/store"
)

type NotesService struct {
	noteStore *store.NoteStore
}

func NewNotesService(noteStore *store.NoteStore) notes.Service {
	return &NotesService{noteStore: noteStore}
}

func (s *NotesService) Create(ctx context.Context, p *notes.CreatePayload) (res *notes.Note, err error) {
	if p.Note == nil {
		return nil, fmt.Errorf("note payload required")
	}

	title := ""
	if p.Note.Title != nil {
		title = *p.Note.Title
	}

	var bodyType *string
	bodyText := ""

	if p.Note.Body != nil {
		if p.Note.Body.List != nil {
			t := "list"
			bodyType = &t
		} else if p.Note.Body.Text != nil {
			t := "text"
			bodyType = &t
			if p.Note.Body.Text.Text != nil {
				bodyText = *p.Note.Body.Text.Text
			}
		}
	}

	color := "DEFAULT"
	if p.Note.Color != nil {
		color = string(*p.Note.Color)
	}

	pinned := false
	if p.Note.Pinned != nil {
		pinned = *p.Note.Pinned
	}

	archived := false
	if p.Note.Archived != nil {
		archived = *p.Note.Archived
	}

	var listItems []*notes.ListItem
	if p.Note.Body != nil && p.Note.Body.List != nil {
		listItems = p.Note.Body.List.ListItems
	}

	return s.noteStore.Create(ctx, title, bodyType, bodyText, color, pinned, archived, p.Note.Labels, listItems)
}

func (s *NotesService) Get(ctx context.Context, p *notes.GetPayload) (res *notes.Note, err error) {
	return s.noteStore.GetByName(ctx, "notes/"+p.ID)
}

func (s *NotesService) List(ctx context.Context, p *notes.ListPayload) (res *notes.ListNotesResponse, err error) {
	return s.noteStore.List(ctx, p.PageSize, p.PageToken, p.Filter)
}

func (s *NotesService) Update(ctx context.Context, p *notes.UpdatePayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}

	return s.noteStore.Update(ctx, id,
		p.Note.Title,
		nil, nil,
		nil, p.Note.Pinned, p.Note.Archived,
		p.Note.Labels, nil,
	)
}

func (s *NotesService) Delete(ctx context.Context, p *notes.DeletePayload) (err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.Delete(ctx, id)
}

func (s *NotesService) Pin(ctx context.Context, p *notes.PinPayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetPinned(ctx, id, true)
}

func (s *NotesService) Unpin(ctx context.Context, p *notes.UnpinPayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetPinned(ctx, id, false)
}

func (s *NotesService) Archive(ctx context.Context, p *notes.ArchivePayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetArchived(ctx, id, true)
}

func (s *NotesService) Unarchive(ctx context.Context, p *notes.UnarchivePayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetArchived(ctx, id, false)
}

func (s *NotesService) Trash(ctx context.Context, p *notes.TrashPayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetTrashed(ctx, id, true)
}

func (s *NotesService) Restore(ctx context.Context, p *notes.RestorePayload) (res *notes.Note, err error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetTrashed(ctx, id, false)
}

func parseNoteUUID(name string) (uuid.UUID, error) {
	name = strings.TrimPrefix(name, "notes/")
	return uuid.Parse(name)
}
