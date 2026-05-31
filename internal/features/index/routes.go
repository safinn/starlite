package index

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"starlite/internal/config"
	"starlite/internal/features/index/services"

	"github.com/alexedwards/scs/v2"
	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/delaneyj/toolbelt/embeddednats"
)

func SetupRoutes(
	ctx context.Context,
	cfg config.Config,
	mux *http.ServeMux,
	sessionManager *scs.SessionManager,
	log *slog.Logger,
	db *toolbeltdb.Database,
	ns *embeddednats.Server,
) error {
	nc, err := ns.Client()
	if err != nil {
		return fmt.Errorf("error creating nats client: %w", err)
	}
	reactionsService := services.NewReactionsService(log, db, nc)

	err = reactionsService.Start(ctx)
	if err != nil {
		return err
	}

	handlers := NewHandlers(log, cfg, reactionsService, sessionManager)

	mux.HandleFunc("GET /", handlers.IndexPage)
	mux.HandleFunc("POST /react", handlers.HandleReaction)
	mux.HandleFunc("GET /react", handlers.ReactionsSSE)

	return nil
}
