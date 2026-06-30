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

func bodyFields(body *notes.Section) (bodyType *string, bodyText *string, listItems []*notes.ListItem) {
	if body == nil {
		return nil, nil, nil
	}
	if body.List != nil {
		t := "list"
		empty := ""
		return &t, &empty, body.List.ListItems
	}
	if body.Text != nil {
		t := "text"
		if body.Text.Text != nil {
			return &t, body.Text.Text, nil
		}
		empty := ""
		return &t, &empty, nil
	}
	return nil, nil, nil
}

func (s *NotesService) Create(ctx context.Context, p *notes.CreatePayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	if p.Note == nil {
		return nil, fmt.Errorf("note payload required")
	}

	title := ""
	if p.Note.Title != nil {
		title = *p.Note.Title
	}

	bodyType, bodyText, listItems := bodyFields(p.Note.Body)

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

	var bodyTextVal string
	if bodyText != nil {
		bodyTextVal = *bodyText
	}
	return s.noteStore.Create(ctx, owner, title, bodyType, bodyTextVal, color, pinned, archived, p.Note.Labels, listItems)
}

func (s *NotesService) Get(ctx context.Context, p *notes.GetPayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)
	return s.noteStore.GetByName(ctx, owner, "notes/"+p.ID)
}

func (s *NotesService) List(ctx context.Context, p *notes.ListPayload) (res *notes.ListNotesResponse, err error) {
	owner := store.OwnerFromContext(ctx)
	return s.noteStore.List(ctx, owner, p.PageSize, p.PageToken, p.Filter, p.Search)
}

func (s *NotesService) Update(ctx context.Context, p *notes.UpdatePayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}

	var color *string
	if p.Note.Color != nil {
		c := string(*p.Note.Color)
		color = &c
	}

	bodyType, bodyText, listItems := bodyFields(p.Note.Body)

	return s.noteStore.Update(ctx, owner, id,
		p.Note.Title,
		bodyType, bodyText,
		color, p.Note.Pinned, p.Note.Archived,
		p.Note.Labels, listItems,
	)
}

func (s *NotesService) Delete(ctx context.Context, p *notes.DeletePayload) (err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.Delete(ctx, owner, id)
}

func (s *NotesService) Pin(ctx context.Context, p *notes.PinPayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetPinned(ctx, owner, id, true)
}

func (s *NotesService) Unpin(ctx context.Context, p *notes.UnpinPayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetPinned(ctx, owner, id, false)
}

func (s *NotesService) Archive(ctx context.Context, p *notes.ArchivePayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetArchived(ctx, owner, id, true)
}

func (s *NotesService) Unarchive(ctx context.Context, p *notes.UnarchivePayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetArchived(ctx, owner, id, false)
}

func (s *NotesService) Trash(ctx context.Context, p *notes.TrashPayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetTrashed(ctx, owner, id, true)
}

func (s *NotesService) Restore(ctx context.Context, p *notes.RestorePayload) (res *notes.Note, err error) {
	owner := store.OwnerFromContext(ctx)

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}
	return s.noteStore.SetTrashed(ctx, owner, id, false)
}

func parseNoteUUID(name string) (uuid.UUID, error) {
	name = strings.TrimPrefix(name, "notes/")
	return uuid.Parse(name)
}
