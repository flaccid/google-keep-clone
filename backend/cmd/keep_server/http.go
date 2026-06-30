package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

//go:embed openapi3.yaml
var openapiSpec []byte

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
		sf := sanitizingFormatter
		mediaServer = mediasvr.New(mediaEndpoints, mux, dec, enc, eh, sf)
		labelsServer = labelssvr.New(labelsEndpoints, mux, dec, enc, eh, sf)
		notesServer = notessvr.New(notesEndpoints, mux, dec, enc, eh, sf)
		permissionsServer = permissionssvr.New(permissionsEndpoints, mux, dec, enc, eh, sf)
	}

	mediasvr.Mount(mux, mediaServer)
	labelssvr.Mount(mux, labelsServer)
	notessvr.Mount(mux, notesServer)
	permissionssvr.Mount(mux, permissionsServer)

	if attachmentStore != nil {
		mux.Handle("POST", "/v1/notes/{noteId}/attachments", uploadHandler(attachmentStore, mux))
	}

	mux.Handle("GET", "/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		spec := string(openapiSpec)
		spec = strings.ReplaceAll(spec, "\n    - url: http://localhost:8080\n", "\n    - url: /v1\n")
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(spec))
	})

	mux.Handle("GET", "/openapi", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>API Docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" })</script>
</body>
</html>`)
	})

	var handler http.Handler = mux
	if dbg {
		handler = debug.HTTP()(handler)
	}
	handler = log.HTTP(ctx)(handler)
	handler = defaultOwner(handler)
	handler = securityHeaders(handler)

	srv := &http.Server{
		Addr:              u.Host,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
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
	log.Printf(ctx, "HTTP %q mounted on %s %s", "GET", "GET", "/openapi")
	log.Printf(ctx, "HTTP %q mounted on %s %s", "GET", "GET", "/openapi.yaml")

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

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data:; font-src 'self' https://cdn.jsdelivr.net; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		next.ServeHTTP(w, r)
	})
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
		if contentType == "" || len(contentType) > 256 {
			contentType = "application/octet-stream"
		}
		allowed := map[string]bool{
			"image/jpeg": true, "image/png": true, "image/gif": true,
			"image/webp": true, "image/svg+xml": true, "application/pdf": true,
			"text/plain": true, "application/octet-stream": true,
		}
		if !allowed[contentType] {
			http.Error(w, `{"error":"unsupported content type"}`, http.StatusUnsupportedMediaType)
			return
		}
		data, err := io.ReadAll(io.LimitReader(r.Body, 32<<20))
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		att, err := attachStore.Upload(r.Context(), nid, contentType, data)
		if err != nil {
			http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
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

func sanitizingFormatter(ctx context.Context, err error) goahttp.Statuser {
	resp := goahttp.NewErrorResponse(ctx, err)
	if se, ok := resp.(*goahttp.ErrorResponse); ok && se.Fault {
		se.Message = "internal server error"
	}
	return resp
}

func defaultOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner := store.OwnerFromContext(r.Context())
		if owner == "" {
			r = r.WithContext(store.WithOwner(r.Context(), "00000000-0000-0000-0000-000000000000"))
		}
		next.ServeHTTP(w, r)
	})
}
