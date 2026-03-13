package index

import (
	"context"
	"log/slog"
	"net/http"

	"starlite/internal/features/index/services"

	"github.com/alexedwards/scs/v2"
	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/delaneyj/toolbelt/embeddednats"
)

func SetupRoutes(
	ctx context.Context,
	mux *http.ServeMux,
	sessionManager *scs.SessionManager,
	log *slog.Logger,
	db *toolbeltdb.Database,
	ns *embeddednats.Server,
) error {
	reactionsService, err := services.NewReactionsService(log, db, ns)
	if err != nil {
		return err
	}

	err = reactionsService.Start(ctx)
	if err != nil {
		return err
	}

	handlers := NewHandlers(log, reactionsService, sessionManager)

	mux.HandleFunc("GET /", handlers.IndexPage)
	mux.HandleFunc("POST /react", handlers.HandleReaction)
	mux.HandleFunc("GET /react", handlers.ReactionsSSE)

	return nil
}
