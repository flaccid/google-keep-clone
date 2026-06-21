package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"

	labelssvr "github.com/flaccid/google-keep-clone/backend/gen/http/labels/server"
	mediasvr "github.com/flaccid/google-keep-clone/backend/gen/http/media/server"
	notessvr "github.com/flaccid/google-keep-clone/backend/gen/http/notes/server"
	permissionssvr "github.com/flaccid/google-keep-clone/backend/gen/http/permissions/server"
	labels "github.com/flaccid/google-keep-clone/backend/gen/labels"
	media "github.com/flaccid/google-keep-clone/backend/gen/media"
	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
	permissions "github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"github.com/flaccid/google-keep-clone/backend/store"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

func handleHTTPServer(ctx context.Context, u *url.URL, mediaEndpoints *media.Endpoints, labelsEndpoints *labels.Endpoints, notesEndpoints *notes.Endpoints, permissionsEndpoints *permissions.Endpoints, wg *sync.WaitGroup, errc chan error, dbg bool, attachmentStore *store.AttachmentStore) {
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
		if dbg {
			debug.MountPprofHandlers(debug.Adapt(mux))
			debug.MountDebugLogEnabler(debug.Adapt(mux))
		}
	}

	var (
		mediaServer       *mediasvr.Server
		labelsServer      *labelssvr.Server
		notesServer       *notessvr.Server
		permissionsServer *permissionssvr.Server
	)
	{
		eh := errorHandler(ctx)
		mediaServer = mediasvr.New(mediaEndpoints, mux, dec, enc, eh, nil)
		labelsServer = labelssvr.New(labelsEndpoints, mux, dec, enc, eh, nil)
		notesServer = notessvr.New(notesEndpoints, mux, dec, enc, eh, nil)
		permissionsServer = permissionssvr.New(permissionsEndpoints, mux, dec, enc, eh, nil)
	}

	mediasvr.Mount(mux, mediaServer)
	labelssvr.Mount(mux, labelsServer)
	notessvr.Mount(mux, notesServer)
	permissionssvr.Mount(mux, permissionsServer)

	if attachmentStore != nil {
		mux.Handle("POST", "/v1/notes/{noteId}/attachments", uploadHandler(attachmentStore, mux))
	}

	var handler http.Handler = mux
	if dbg {
		handler = debug.HTTP()(handler)
	}
	handler = log.HTTP(ctx)(handler)

	srv := &http.Server{Addr: u.Host, Handler: handler, ReadHeaderTimeout: time.Second * 60}
	for _, m := range mediaServer.Mounts {
		log.Printf(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range labelsServer.Mounts {
		log.Printf(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range notesServer.Mounts {
		log.Printf(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range permissionsServer.Mounts {
		log.Printf(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	log.Printf(ctx, "HTTP %q mounted on %s %s", "POST", "POST", "/v1/notes/{noteId}/attachments")

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		go func() {
			log.Printf(ctx, "HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		log.Printf(ctx, "shutting down HTTP server at %q", u.Host)

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf(shutdownCtx, "failed to shutdown: %v", err)
		}
	}()
}

func uploadHandler(attachStore *store.AttachmentStore, mux goahttp.Muxer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noteID := mux.Vars(r)["noteId"]

		nid, err := uuid.Parse(noteID)
		if err != nil {
			http.Error(w, `{"error":"invalid note ID"}`, http.StatusBadRequest)
			return
		}

		contentType := r.Header.Get("Content-Type")
		data, err := io.ReadAll(io.LimitReader(r.Body, 32<<20))
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		att, err := attachStore.Upload(r.Context(), nid, contentType, data)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(att)
	}
}

func errorHandler(logCtx context.Context) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		log.Printf(logCtx, "ERROR: %s", err.Error())
	}
}
