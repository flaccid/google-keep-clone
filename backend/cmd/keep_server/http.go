package main

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	labelssvr "github.com/flaccid/google-keep-clone/backend/gen/http/labels/server"
	mediasvr "github.com/flaccid/google-keep-clone/backend/gen/http/media/server"
	notessvr "github.com/flaccid/google-keep-clone/backend/gen/http/notes/server"
	permissionssvr "github.com/flaccid/google-keep-clone/backend/gen/http/permissions/server"
	labels "github.com/flaccid/google-keep-clone/backend/gen/labels"
	media "github.com/flaccid/google-keep-clone/backend/gen/media"
	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
	permissions "github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, mediaEndpoints *media.Endpoints, labelsEndpoints *labels.Endpoints, notesEndpoints *notes.Endpoints, permissionsEndpoints *permissions.Endpoints, wg *sync.WaitGroup, errc chan error, dbg bool) {

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/implement/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and mount debug and profiler
	// endpoints in debug mode.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
		if dbg {
			// Mount pprof handlers for memory profiling under /debug/pprof.
			debug.MountPprofHandlers(debug.Adapt(mux))
			// Mount /debug endpoint to enable or disable debug logs at runtime.
			debug.MountDebugLogEnabler(debug.Adapt(mux))
		}
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
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

	// Configure the mux.
	mediasvr.Mount(mux, mediaServer)
	labelssvr.Mount(mux, labelsServer)
	notessvr.Mount(mux, notesServer)
	permissionssvr.Mount(mux, permissionsServer)

	var handler http.Handler = mux
	if dbg {
		// Log query and response bodies if debug logs are enabled.
		handler = debug.HTTP()(handler)
	}
	handler = log.HTTP(ctx)(handler)

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
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

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			log.Printf(ctx, "HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		log.Printf(ctx, "shutting down HTTP server at %q", u.Host)

		// Shutdown gracefully with a 30s timeout.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf(shutdownCtx, "failed to shutdown: %v", err)
		}
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logCtx context.Context) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		log.Printf(logCtx, "ERROR: %s", err.Error())
	}
}
