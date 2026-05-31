package reactions

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"starlite/internal/features/reactions/stream"

	"github.com/alexedwards/scs/v2"
	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/delaneyj/toolbelt/embeddednats"
)

func SetupRoutes(
	ctx context.Context,
	isDev bool,
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

	reactionStream, err := stream.New(ctx, log, db, nc)
	if err != nil {
		return err
	}

	handlers := NewHandlers(log, isDev, reactionStream, sessionManager)

	mux.HandleFunc("GET /", handlers.IndexPage)
	mux.HandleFunc("POST /react", handlers.HandleReaction)
	mux.HandleFunc("GET /react", handlers.ReactionsSSE)

	return nil
}
