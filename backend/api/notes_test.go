package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goahttp "goa.design/goa/v3/http"

	"github.com/flaccid/google-keep-clone/backend/api"
	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
	notessvr "github.com/flaccid/google-keep-clone/backend/gen/http/notes/server"
	"github.com/flaccid/google-keep-clone/backend/store"
)

func newTestServer(t *testing.T) (*httptest.Server, *store.NoteStore) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://keep:keep@localhost:5432/keep?sslmode=disable"
	}
	if err := store.RunMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := store.Connect(context.Background())
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	truncateTables(t, pool)

	noteStore := store.NewNoteStore(pool)
	svc := api.NewNotesService(noteStore)
	endpoints := notes.NewEndpoints(svc)

	mux := goahttp.NewMuxer()
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	eh := func(ctx context.Context, w http.ResponseWriter, err error) {}

	notesServer := notessvr.New(endpoints, mux, dec, enc, eh, nil)
	notessvr.Mount(mux, notesServer)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, noteStore
}

func truncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE note_labels, list_items, permissions, notes, labels CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func TestAPICreateTextNote(t *testing.T) {
	server, _ := newTestServer(t)

	body := `{"title":"API Note","body":{"text":{"text":"Hello from API"}}}`
	resp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Contains(t, result["name"], "notes/")
	assert.Equal(t, "API Note", result["title"])
	assert.Equal(t, "Hello from API", result["body"].(map[string]any)["text"].(map[string]any)["text"])
}

func TestAPICreateNoteWithList(t *testing.T) {
	server, _ := newTestServer(t)

	body := `{
		"title":"Checklist",
		"body":{
			"list":{
				"listItems":[
					{"text":{"text":"Item A"},"checked":false},
					{"text":{"text":"Item B"},"checked":true},
					{"text":{"text":"Parent"},"checked":false,"childListItems":[
						{"text":{"text":"Child"},"checked":false}
					]}
				]
			}
		}
	}`
	resp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "Checklist", result["title"])

	list := result["body"].(map[string]any)["list"].(map[string]any)
	items := list["listItems"].([]any)
	assert.Len(t, items, 3)

	item0 := items[0].(map[string]any)
	assert.Equal(t, "Item A", item0["text"].(map[string]any)["text"])
	assert.Equal(t, false, item0["checked"])

	item1 := items[1].(map[string]any)
	assert.Equal(t, true, item1["checked"])

	item2 := items[2].(map[string]any)
	children := item2["childListItems"].([]any)
	assert.Len(t, children, 1)
}

func TestAPIGetNote(t *testing.T) {
	server, _ := newTestServer(t)

	// Create a note
	createBody := `{"title":"Get Me","body":{"text":{"text":"Content"}}}`
	createResp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(createBody)))
	require.NoError(t, err)
	defer createResp.Body.Close()

	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	name := created["name"].(string)
	id := strings.TrimPrefix(name, "notes/")

	// Get the note
	getResp, err := http.Get(server.URL + "/v1/notes/" + id)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&got))
	assert.Equal(t, name, got["name"])
	assert.Equal(t, "Get Me", got["title"])
}

func TestAPIListNotes(t *testing.T) {
	server, _ := newTestServer(t)

	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"title":"Note %d","body":{"text":{"text":"Content"}}}`, i)
		resp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		resp.Body.Close()
	}

	resp, err := http.Get(server.URL + "/v1/notes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	notes_ := result["notes"].([]any)
	assert.GreaterOrEqual(t, len(notes_), 3)
}

func TestAPIUpdateNote(t *testing.T) {
	server, _ := newTestServer(t)

	createBody := `{"title":"Original","body":{"text":{"text":"Old content"}}}`
	createResp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(createBody)))
	require.NoError(t, err)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	createResp.Body.Close()
	id := strings.TrimPrefix(created["name"].(string), "notes/")

	t.Run("update title", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", server.URL+"/v1/notes/"+id, bytes.NewReader([]byte(`{"title":"Updated"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, "Updated", result["title"])
	})

	t.Run("update to list body", func(t *testing.T) {
		patchBody := `{
			"body":{
				"list":{
					"listItems":[
						{"text":{"text":"New Item"},"checked":false}
					]
				}
			}
		}`
		req, _ := http.NewRequest("PATCH", server.URL+"/v1/notes/"+id, bytes.NewReader([]byte(patchBody)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		list := result["body"].(map[string]any)["list"].(map[string]any)
		items := list["listItems"].([]any)
		assert.Len(t, items, 1)
		assert.Equal(t, "New Item", items[0].(map[string]any)["text"].(map[string]any)["text"])
		assert.Equal(t, "Updated", result["title"])
	})
}

func TestAPIDeleteNote(t *testing.T) {
	server, _ := newTestServer(t)

	createBody := `{"title":"Delete Me","body":{"text":{"text":"Gone"}}}`
	createResp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(createBody)))
	require.NoError(t, err)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	createResp.Body.Close()
	id := strings.TrimPrefix(created["name"].(string), "notes/")

	req, _ := http.NewRequest("DELETE", server.URL+"/v1/notes/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAPIPinUnpin(t *testing.T) {
	server, _ := newTestServer(t)

	createBody := `{"title":"Pin Me","body":{"text":{"text":"Test"}}}`
	createResp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(createBody)))
	require.NoError(t, err)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	createResp.Body.Close()
	id := strings.TrimPrefix(created["name"].(string), "notes/")

	pinResp, err := http.Post(server.URL+"/v1/notes/"+id+":pin", "application/json", nil)
	require.NoError(t, err)
	defer pinResp.Body.Close()
	assert.Equal(t, http.StatusOK, pinResp.StatusCode)
	var pinned map[string]any
	require.NoError(t, json.NewDecoder(pinResp.Body).Decode(&pinned))
	assert.Equal(t, true, pinned["pinned"])
}

func TestAPIArchiveUnarchive(t *testing.T) {
	server, _ := newTestServer(t)

	createBody := `{"title":"Archive Me","body":{"text":{"text":"Test"}}}`
	createResp, err := http.Post(server.URL+"/v1/notes", "application/json", bytes.NewReader([]byte(createBody)))
	require.NoError(t, err)
	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	createResp.Body.Close()
	id := strings.TrimPrefix(created["name"].(string), "notes/")

	archiveResp, err := http.Post(server.URL+"/v1/notes/"+id+":archive", "application/json", nil)
	require.NoError(t, err)
	defer archiveResp.Body.Close()
	assert.Equal(t, http.StatusOK, archiveResp.StatusCode)
	var archived map[string]any
	require.NoError(t, json.NewDecoder(archiveResp.Body).Decode(&archived))
	assert.Equal(t, true, archived["archived"])
}
