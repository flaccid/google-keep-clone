package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/flaccid/google-keep-clone/backend/api"
	genhttp "github.com/flaccid/google-keep-clone/backend/gen/http"
	mediasvr "github.com/flaccid/google-keep-clone/backend/gen/http/media/server"
	notessvr "github.com/flaccid/google-keep-clone/backend/gen/http/notes/server"
	permissionssvr "github.com/flaccid/google-keep-clone/backend/gen/http/permissions/server"
	"github.com/flaccid/google-keep-clone/backend/gen/notes"
	"github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"github.com/flaccid/google-keep-clone/backend/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	goahttp "goa.design/goa/v3/http"
)

func TestGeneratedMediaAndSpecCompatibility(t *testing.T) {
	for _, body := range [][]byte{[]byte(`{"name":"synthetic attachment"}`), {0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}} {
		w := httptest.NewRecorder()
		err := mediasvr.EncodeDownloadResponse(goahttp.ResponseEncoder)(context.Background(), w, body)
		require.NoError(t, err)
		require.Equal(t, body, w.Body.Bytes())
		if body[0] == '{' {
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))
		} else {
			require.Equal(t, "image/png", w.Header().Get("Content-Type"))
		}
	}
	require.True(t, json.Valid(genhttp.Spec))
	require.True(t, bytes.Contains(genhttp.Spec, []byte("READER")))
}

func TestPermissionServiceRejectsMissingBatch(t *testing.T) {
	service := api.NewPermissionsService(nil)
	for _, payload := range []*permissions.BatchCreatePayload{nil, {}} {
		result, err := service.BatchCreate(context.Background(), payload)
		require.Nil(t, result)
		var invalid permissions.InvalidRequest
		require.ErrorAs(t, err, &invalid)
	}
	for _, payload := range []*permissions.BatchDeletePayload{nil, {}} {
		err := service.BatchDelete(context.Background(), payload)
		var invalid permissions.InvalidRequest
		require.ErrorAs(t, err, &invalid)
	}
}

func TestPermissionHTTPContract(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn, "DATABASE_URL must identify the disposable test database")
	require.NoError(t, store.RunMigrations(dsn))
	ctx := context.Background()
	pool, err := store.Connect(ctx)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	truncateTables(t, pool)
	owner, peer := uuid.NewString(), uuid.NewString()
	noteStore := store.NewNoteStore(pool, nil)
	note, err := noteStore.Create(ctx, owner, "permission HTTP fixture", nil, "", "DEFAULT", false, false, nil, nil)
	require.NoError(t, err)
	noteID := strings.TrimPrefix(*note.Name, "notes/")
	mux := goahttp.NewMuxer()
	service := api.NewPermissionsService(store.NewPermissionStore(pool))
	permissionServer := permissionssvr.New(permissions.NewEndpoints(service), mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil)
	permissionssvr.Mount(mux, permissionServer)
	notesServer := notessvr.New(notes.NewEndpoints(api.NewNotesService(noteStore)), mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil)
	notessvr.Mount(mux, notesServer)
	request := func(actor, method, route, body string) *httptest.ResponseRecorder {
		t.Helper()
		caller := peer
		if actor == "owner" {
			caller = owner
		}
		requestCtx := store.WithEmail(store.WithOwner(ctx, caller), actor+"@example.invalid")
		req := httptest.NewRequest(method, "/v1/notes/"+noteID+route, strings.NewReader(body)).WithContext(requestCtx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	grant := func(email, role string) string {
		t.Helper()
		w := request("owner", "POST", "/permissions:batchCreate", `{"requests":[{"email":"`+email+`","role":"`+role+`"}]}`)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var result []struct {
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.Len(t, result, 1)
		return result[0].Name
	}
	for _, role := range []string{"WRITER", "READER", "OWNER"} {
		grant(strings.ToLower(role)+"@example.invalid", role)
	}
	for _, actor := range []string{"writer", "reader", "stranger", "owner-impostor"} {
		if actor == "owner-impostor" {
			grant(actor+"@example.invalid", "OWNER")
		}
		w := request(actor, "POST", "/permissions:batchCreate", `{"requests":[{"email":"new@example.invalid","role":"WRITER"}]}`)
		require.Equal(t, http.StatusForbidden, w.Code)
		w = request(actor, "POST", "/permissions:batchDelete", `{"names":[]}`)
		require.Equal(t, http.StatusForbidden, w.Code)
	}
	w := request("reader", "GET", "", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	w = request("reader", "PATCH", "", `{"title":"forbidden reader edit"}`)
	require.NotContains(t, []int{200, 201, 204}, w.Code)
	w = request("reader", "GET", "", "")
	require.NotContains(t, w.Body.String(), "forbidden reader edit")
	first := grant("reader@example.invalid", "WRITER")
	second := grant("reader@example.invalid", "READER")
	require.Equal(t, first, second)
	w = request("owner", "POST", "/permissions:batchCreate", `{"requests":[{"email":"duplicate@example.invalid","role":"WRITER"},{"email":"duplicate@example.invalid","role":"READER"}]}`)
	require.Equal(t, http.StatusBadRequest, w.Code)
	w = request("owner", "POST", "/permissions:batchDelete", `{"names":["`+first+`","notes/`+uuid.NewString()+`/permissions/`+uuid.NewString()+`"]}`)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, http.StatusOK, request("reader", "GET", "", "").Code)
	w = request("owner", "POST", "/permissions:batchDelete", `{"names":["`+second+`"]}`)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, http.StatusNoContent, request("owner", "POST", "/permissions:batchDelete", `{"names":["`+second+`"]}`).Code)
	w = request("reader", "GET", "", "")
	require.NotEqual(t, http.StatusOK, w.Code)
	existingNoteID := noteID
	noteID = uuid.NewString()
	require.Equal(t, http.StatusNotFound, request("owner", "POST", "/permissions:batchCreate", `{"requests":[]}`).Code)
	noteID = existingNoteID
	pool.Close()
	w = request("owner", "POST", "/permissions:batchCreate", `{"requests":[]}`)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.NotContains(t, w.Body.String(), "closed")
	require.NotContains(t, w.Body.String(), "postgres")
}
