package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"starlite/internal/config"
	"starlite/internal/features/index"
	"starlite/internal/static"

	"github.com/alexedwards/scs/v2"
	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/starfederation/datastar-go/datastar"
)

// SetupRoutes registers all the application routes on the provided mux.
func SetupRoutes(
	ctx context.Context,
	mux *http.ServeMux,
	sessionManager *scs.SessionManager,
	sqliteDB *toolbeltdb.Database,
	ns *embeddednats.Server,
	log *slog.Logger,
) error {
	if config.Global.Env == config.Dev {
		setupReload(mux, log)
	}

	mux.Handle("GET /static/", static.Handler(log))

	if err := errors.Join(
		index.SetupRoutes(ctx, mux, sessionManager, log, sqliteDB, ns),
	); err != nil {
		return fmt.Errorf("error setting up routes: %w", err)
	}

	return nil
}

// setupReload registers the /reload and /hotreload endpoints on the provided mux.
// These endpoints are used for live reloading the browser during development.
func setupReload(mux *http.ServeMux, log *slog.Logger) {
	var (
		mu        sync.Mutex
		listeners = make(map[chan struct{}]struct{})
	)

	// Broadcast to all connected SSE clients (for CSS/static file changes)
	broadcast := func() {
		mu.Lock()
		defer mu.Unlock()
		for ch := range listeners {
			select {
			case ch <- struct{}{}:
			default: // Skip if channel is full
			}
		}
	}

	var hotReloadOnce sync.Once

	mux.HandleFunc("GET /reload", func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		reload := func() {
			err := sse.ExecuteScript("window.location.reload()")
			if err != nil {
				log.Error("error executing reload script", "error", err)
			}
		}

		// Immediate reload when connecting (server restart = new Go code deployed)
		hotReloadOnce.Do(reload)

		// Register this SSE connection for future broadcasts
		ch := make(chan struct{}, 1)
		mu.Lock()
		listeners[ch] = struct{}{}
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(listeners, ch)
			mu.Unlock()
		}()

		// Wait for broadcast signal or client disconnect
		select {
		case <-ch:
			reload()
		case <-r.Context().Done():
			return
		}
	})

	mux.HandleFunc("GET /hotreload", func(w http.ResponseWriter, r *http.Request) {
		broadcast()
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Error("error writing hotreload response", "error", err)
		}
	})
}
