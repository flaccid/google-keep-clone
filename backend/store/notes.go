package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
)

type NoteStore struct {
	pool *pgxpool.Pool
}

type noteRow struct {
	ID        uuid.UUID
	Title     string
	BodyType  *string
	BodyText  string
	Color     string
	Pinned    bool
	Archived  bool
	Trashed   bool
	TrashTime *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type listItemRow struct {
	ID        uuid.UUID
	NoteID    uuid.UUID
	ParentID  *uuid.UUID
	Text      string
	Checked   bool
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type labelRow struct {
	ID          uuid.UUID
	DisplayName string
}

type permissionRow struct {
	ID     uuid.UUID
	NoteID uuid.UUID
	Email  string
	Role   string
}

func NewNoteStore(pool *pgxpool.Pool) *NoteStore {
	return &NoteStore{pool: pool}
}

func (s *NoteStore) Create(ctx context.Context, title string, bodyType *string, bodyText string, color string, pinned bool, archived bool, labels []string, listItems []*notes.ListItem) (*notes.Note, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	id := uuid.New()
	now := time.Now().UTC()

	_, err = tx.Exec(ctx, `
		INSERT INTO notes (id, title, body_type, body_text, color, pinned, archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, title, bodyType, bodyText, color, pinned, archived, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert note: %w", err)
	}

	if err := s.insertListItems(ctx, tx, id, nil, listItems); err != nil {
		return nil, err
	}

	for _, labelName := range labels {
		if err := s.ensureLabelOnNote(ctx, tx, id, labelName); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.GetByID(ctx, id)
}

func (s *NoteStore) GetByID(ctx context.Context, id uuid.UUID) (*notes.Note, error) {
	row, err := s.queryNote(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.assembleNote(ctx, row)
}

func (s *NoteStore) GetByName(ctx context.Context, name string) (*notes.Note, error) {
	id, err := parseNoteName(name)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *NoteStore) List(ctx context.Context, pageSize *int, pageToken *string, filter *string, search *string) (*notes.ListNotesResponse, error) {
	limit := 20
	if pageSize != nil && *pageSize > 0 {
		limit = *pageSize
	}

	var offset int
	if pageToken != nil && *pageToken != "" {
		fmt.Sscanf(*pageToken, "%d", &offset)
	}

	var conditions []string
	var args []any

	if filter != nil && *filter != "" {
		safe, err := safeFilter(*filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter: %w", err)
		}
		conditions = append(conditions, safe)
	} else {
		conditions = append(conditions, "trashed = false")
	}

	argIdx := 0
	if search != nil && *search != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`(title ILIKE $%d OR body_text ILIKE $%d)`, argIdx, argIdx))
		args = append(args, "%"+*search+"%")
	}

	argIdx++
	limitArg := argIdx
	argIdx++
	offsetArg := argIdx

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, title, body_type, body_text, color, pinned, archived, trashed, trash_time, created_at, updated_at
		FROM notes WHERE %s
		ORDER BY pinned DESC, updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg)

	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()

	var noteRows []noteRow
	for rows.Next() {
		var nr noteRow
		if err := rows.Scan(&nr.ID, &nr.Title, &nr.BodyType, &nr.BodyText, &nr.Color, &nr.Pinned, &nr.Archived, &nr.Trashed, &nr.TrashTime, &nr.CreatedAt, &nr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		noteRows = append(noteRows, nr)
	}

	notes_ := make([]*notes.Note, 0, len(noteRows))
	for _, nr := range noteRows {
		n, err := s.assembleNote(ctx, nr)
		if err != nil {
			return nil, err
		}
		notes_ = append(notes_, n)
	}

	var nextToken *string
	if len(notes_) == limit {
		t := fmt.Sprintf("%d", offset+limit)
		nextToken = &t
	}

	return &notes.ListNotesResponse{
		Notes:         notes_,
		NextPageToken: nextToken,
	}, nil
}

func (s *NoteStore) Update(ctx context.Context, id uuid.UUID, title *string, bodyType *string, bodyText *string, color *string, pinned *bool, archived *bool, labels []string, listItems []*notes.ListItem) (*notes.Note, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	if title != nil {
		_, err = tx.Exec(ctx, `UPDATE notes SET title = $1, updated_at = $2 WHERE id = $3`, *title, now, id)
		if err != nil {
			return nil, fmt.Errorf("update title: %w", err)
		}
	}

	if bodyType != nil && bodyText != nil {
		_, err = tx.Exec(ctx, `UPDATE notes SET body_type = $1, body_text = $2, updated_at = $3 WHERE id = $4`, *bodyType, *bodyText, now, id)
		if err != nil {
			return nil, fmt.Errorf("update body: %w", err)
		}
	}

	if color != nil {
		_, err = tx.Exec(ctx, `UPDATE notes SET color = $1, updated_at = $2 WHERE id = $3`, *color, now, id)
		if err != nil {
			return nil, fmt.Errorf("update color: %w", err)
		}
	}

	if pinned != nil {
		_, err = tx.Exec(ctx, `UPDATE notes SET pinned = $1, updated_at = $2 WHERE id = $3`, *pinned, now, id)
		if err != nil {
			return nil, fmt.Errorf("update pinned: %w", err)
		}
	}

	if archived != nil {
		_, err = tx.Exec(ctx, `UPDATE notes SET archived = $1, updated_at = $2 WHERE id = $3`, *archived, now, id)
		if err != nil {
			return nil, fmt.Errorf("update archived: %w", err)
		}
	}

	if listItems != nil {
		_, err = tx.Exec(ctx, `DELETE FROM list_items WHERE note_id = $1`, id)
		if err != nil {
			return nil, fmt.Errorf("delete list items: %w", err)
		}
		if err := s.insertListItems(ctx, tx, id, nil, listItems); err != nil {
			return nil, err
		}
	}

	if labels != nil {
		_, err = tx.Exec(ctx, `DELETE FROM note_labels WHERE note_id = $1`, id)
		if err != nil {
			return nil, fmt.Errorf("delete note labels: %w", err)
		}
		for _, labelName := range labels {
			if err := s.ensureLabelOnNote(ctx, tx, id, labelName); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.GetByID(ctx, id)
}

func (s *NoteStore) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) (*notes.Note, error) {
	_, err := s.pool.Exec(ctx, `UPDATE notes SET pinned = $1, updated_at = $2 WHERE id = $3`, pinned, time.Now().UTC(), id)
	if err != nil {
		return nil, fmt.Errorf("set pinned: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *NoteStore) SetArchived(ctx context.Context, id uuid.UUID, archived bool) (*notes.Note, error) {
	_, err := s.pool.Exec(ctx, `UPDATE notes SET archived = $1, updated_at = $2 WHERE id = $3`, archived, time.Now().UTC(), id)
	if err != nil {
		return nil, fmt.Errorf("set archived: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *NoteStore) SetTrashed(ctx context.Context, id uuid.UUID, trashed bool) (*notes.Note, error) {
	var trashTime *time.Time
	if trashed {
		t := time.Now().UTC()
		trashTime = &t
	}
	_, err := s.pool.Exec(ctx, `UPDATE notes SET trashed = $1, trash_time = $2, updated_at = $3 WHERE id = $4`, trashed, trashTime, time.Now().UTC(), id)
	if err != nil {
		return nil, fmt.Errorf("set trashed: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *NoteStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}

func (s *NoteStore) queryNote(ctx context.Context, id uuid.UUID) (noteRow, error) {
	var nr noteRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, body_type, body_text, color, pinned, archived, trashed, trash_time, created_at, updated_at
		FROM notes WHERE id = $1
	`, id).Scan(&nr.ID, &nr.Title, &nr.BodyType, &nr.BodyText, &nr.Color, &nr.Pinned, &nr.Archived, &nr.Trashed, &nr.TrashTime, &nr.CreatedAt, &nr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nr, fmt.Errorf("note not found")
		}
		return nr, fmt.Errorf("query note: %w", err)
	}
	return nr, nil
}

func (s *NoteStore) assembleNote(ctx context.Context, row noteRow) (*notes.Note, error) {
	name := fmt.Sprintf("notes/%s", row.ID.String())
	createTime := row.CreatedAt.Format(time.RFC3339)
	updateTime := row.UpdatedAt.Format(time.RFC3339)

	n := &notes.Note{
		Name:       &name,
		CreateTime: &createTime,
		UpdateTime: &updateTime,
		Title:      &row.Title,
		Trashed:    &row.Trashed,
		Pinned:     &row.Pinned,
		Archived:   &row.Archived,
		Color:      ptr(notes.ColorValue(row.Color)),
	}

	if row.TrashTime != nil {
		t := row.TrashTime.Format(time.RFC3339)
		n.TrashTime = &t
	}

	bodyType := ""
	if row.BodyType != nil {
		bodyType = *row.BodyType
	}

	switch bodyType {
	case "list":
		items, err := s.getListItems(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		n.Body = &notes.Section{
			List: &notes.ListContent{
				ListItems: items,
			},
		}
	default:
		n.Body = &notes.Section{
			Text: &notes.TextContent{
				Text: &row.BodyText,
			},
		}
	}

	labels, err := s.getLabels(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	n.Labels = labels

	return n, nil
}

func (s *NoteStore) getListItems(ctx context.Context, noteID uuid.UUID) ([]*notes.ListItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, text, checked, sort_order, parent_id
		FROM list_items WHERE note_id = $1
		ORDER BY sort_order
	`, noteID)
	if err != nil {
		return nil, fmt.Errorf("query list items: %w", err)
	}
	defer rows.Close()

	var allItems []listItemRow
	for rows.Next() {
		var li listItemRow
		if err := rows.Scan(&li.ID, &li.Text, &li.Checked, &li.SortOrder, &li.ParentID); err != nil {
			return nil, fmt.Errorf("scan list item: %w", err)
		}
		allItems = append(allItems, li)
	}

	itemMap := make(map[uuid.UUID]*notes.ListItem)
	var rootItems []*notes.ListItem

	for _, li := range allItems {
		text := li.Text
		checked := li.Checked
		item := &notes.ListItem{
			Text:    &notes.TextContent{Text: &text},
			Checked: &checked,
		}
		itemMap[li.ID] = item
	}

	for _, li := range allItems {
		item := itemMap[li.ID]
		if li.ParentID != nil {
			if parent, ok := itemMap[*li.ParentID]; ok {
				parent.ChildListItems = append(parent.ChildListItems, item)
			}
		} else {
			rootItems = append(rootItems, item)
		}
	}

	return rootItems, nil
}

func (s *NoteStore) getLabels(ctx context.Context, noteID uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT l.display_name FROM labels l
		JOIN note_labels nl ON nl.label_id = l.id
		WHERE nl.note_id = $1
		ORDER BY l.display_name
	`, noteID)
	if err != nil {
		return nil, fmt.Errorf("query labels: %w", err)
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		labels = append(labels, name)
	}
	return labels, nil
}

func (s *NoteStore) insertListItems(ctx context.Context, tx pgx.Tx, noteID uuid.UUID, parentID *uuid.UUID, items []*notes.ListItem) error {
	for i, item := range items {
		id := uuid.New()
		text := ""
		if item.Text != nil && item.Text.Text != nil {
			text = *item.Text.Text
		}
		checked := false
		if item.Checked != nil {
			checked = *item.Checked
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO list_items (id, note_id, parent_id, text, checked, sort_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, id, noteID, parentID, text, checked, i, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			return fmt.Errorf("insert list item: %w", err)
		}

		if len(item.ChildListItems) > 0 {
			if err := s.insertListItems(ctx, tx, noteID, &id, item.ChildListItems); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *NoteStore) ensureLabelOnNote(ctx context.Context, tx pgx.Tx, noteID uuid.UUID, labelName string) error {
	var labelID uuid.UUID
	err := tx.QueryRow(ctx, `INSERT INTO labels (id, display_name) VALUES ($1, $2) ON CONFLICT (display_name) DO UPDATE SET display_name = EXCLUDED.display_name RETURNING id`, uuid.New(), labelName).Scan(&labelID)
	if err != nil {
		return fmt.Errorf("upsert label: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO note_labels (note_id, label_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, noteID, labelID)
	if err != nil {
		return fmt.Errorf("insert note label: %w", err)
	}

	return nil
}

var allowedFilterColumns = map[string]bool{
	"trashed":  true,
	"archived": true,
	"pinned":   true,
}

func safeFilter(raw string) (string, error) {
	parts := strings.Split(raw, " AND ")
	var safeParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		pairs := strings.SplitN(part, " = ", 2)
		if len(pairs) != 2 {
			return "", fmt.Errorf("malformed condition: %q", part)
		}
		col := strings.TrimSpace(pairs[0])
		val := strings.TrimSpace(pairs[1])
		if !allowedFilterColumns[col] {
			return "", fmt.Errorf("disallowed column: %q", col)
		}
		if val != "true" && val != "false" {
			return "", fmt.Errorf("disallowed value: %q", val)
		}
		safeParts = append(safeParts, fmt.Sprintf("%s = %s", col, val))
	}
	return strings.Join(safeParts, " AND "), nil
}

func parseNoteName(name string) (uuid.UUID, error) {
	if len(name) > 6 && name[:6] == "notes/" {
		name = name[6:]
	}
	return uuid.Parse(name)
}

func ptr[T any](v T) *T {
	return &v
}
