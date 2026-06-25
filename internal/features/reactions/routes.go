package reactions

import (
	"context"
	"log/slog"
	"net/http"

	"starlite/internal/config"
	"starlite/internal/features/reactions/stream"

	"github.com/alexedwards/scs/v2"
	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/nats-io/nats.go"
)

func SetupRoutes(
	ctx context.Context,
	cfg config.Config,
	mux *http.ServeMux,
	sessionManager *scs.SessionManager,
	log *slog.Logger,
	db *toolbeltdb.Database,
	nc *nats.Conn,
) error {
	reactionStream, err := stream.New(ctx, log, db, nc)
	if err != nil {
		return err
	}

	handlers := NewHandlers(log, cfg.IsDev(), cfg.BaseURL, reactionStream, sessionManager)

	mux.HandleFunc("GET /{$}", handlers.IndexPage)
	mux.HandleFunc("POST /react", handlers.HandleReaction)
	mux.HandleFunc("GET /react", handlers.ReactionsSSE)

	return nil
}
