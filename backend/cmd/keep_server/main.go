package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/flaccid/google-keep-clone/backend/api"
	labels "github.com/flaccid/google-keep-clone/backend/gen/labels"
	media "github.com/flaccid/google-keep-clone/backend/gen/media"
	notes "github.com/flaccid/google-keep-clone/backend/gen/notes"
	permissions "github.com/flaccid/google-keep-clone/backend/gen/permissions"
	"github.com/flaccid/google-keep-clone/backend/store"
	"goa.design/clue/debug"
	"goa.design/clue/log"
)

func main() {
	var (
		hostF     = flag.String("host", "0.0.0.0", "Server host (valid values: localhost, 0.0.0.0)")
		domainF   = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if *dbgF {
		ctx = log.Context(ctx, log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf(ctx, fmt.Errorf("DATABASE_URL environment variable is required"), "")
	}

	if err := store.RunMigrations(dsn); err != nil {
		log.Fatalf(ctx, err, "failed to run migrations")
	}

	pool, err := store.Connect(ctx)
	if err != nil {
		log.Fatalf(ctx, err, "failed to connect to database")
	}
	defer pool.Close()

	var (
		attachmentStore *store.AttachmentStore
		mediaSvc        media.Service
		labelsSvc       labels.Service
		notesSvc        notes.Service
		permissionsSvc  permissions.Service
	)
	{
		attachmentStore = store.NewAttachmentStore(pool)
		mediaSvc = api.NewMediaService(attachmentStore)
		labelsSvc = api.NewLabelsService(store.NewLabelStore(pool))
		notesSvc = api.NewNotesService(store.NewNoteStore(pool, attachmentStore))
		permissionsSvc = api.NewPermissionsService(store.NewPermissionStore(pool))
	}

	var (
		mediaEndpoints       *media.Endpoints
		labelsEndpoints      *labels.Endpoints
		notesEndpoints       *notes.Endpoints
		permissionsEndpoints *permissions.Endpoints
	)
	{
		mediaEndpoints = media.NewEndpoints(mediaSvc)
		mediaEndpoints.Use(debug.LogPayloads())
		mediaEndpoints.Use(log.Endpoint)
		labelsEndpoints = labels.NewEndpoints(labelsSvc)
		labelsEndpoints.Use(debug.LogPayloads())
		labelsEndpoints.Use(log.Endpoint)
		notesEndpoints = notes.NewEndpoints(notesSvc)
		notesEndpoints.Use(debug.LogPayloads())
		notesEndpoints.Use(log.Endpoint)
		permissionsEndpoints = permissions.NewEndpoints(permissionsSvc)
		permissionsEndpoints.Use(debug.LogPayloads())
		permissionsEndpoints.Use(log.Endpoint)
	}

	errc := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	switch *hostF {
	case "localhost":
		{
			addr := "http://localhost:8080"
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, mediaEndpoints, labelsEndpoints, notesEndpoints, permissionsEndpoints, &wg, errc, *dbgF, attachmentStore)
		}

	case "0.0.0.0":
		{
			addr := "http://0.0.0.0:8080"
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, mediaEndpoints, labelsEndpoints, notesEndpoints, permissionsEndpoints, &wg, errc, *dbgF, attachmentStore)
		}

	default:
		log.Fatal(ctx, fmt.Errorf("invalid host argument: %q (valid hosts: localhost, 0.0.0.0)", *hostF))
	}

	log.Printf(ctx, "exiting (%v)", <-errc)

	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}
