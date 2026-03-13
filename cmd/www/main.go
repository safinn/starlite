package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"starlite/internal/config"
	"starlite/internal/db"
	"starlite/internal/middleware"
	"starlite/internal/router"
	"starlite/pkg/logger"
	"starlite/pkg/nats"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	log := logger.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, log); err != nil && err != http.ErrServerClosed {
		log.Error("error running server", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	eg, egctx := errgroup.WithContext(ctx)

	db, err := db.SetupDB(egctx, "./", config.Global.Env != config.Prod)
	if err != nil {
		return fmt.Errorf("error setting up database: %w", err)
	}

	ns, err := nats.SetupNATS(egctx, log)
	if err != nil {
		return err
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour * 30
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = config.Global.Env == config.Prod

	mux := http.NewServeMux()

	if err := router.SetupRoutes(egctx, mux, sessionManager, db, ns, log); err != nil {
		return fmt.Errorf("error setting up routes: %w", err)
	}

	handler := middleware.RecoveryMiddleware(
		log,
		middleware.SecurityHeadersMiddleware(
			middleware.LoggingMiddleware(log)(
				sessionManager.LoadAndSave(mux),
			),
		),
	)

	srv := &http.Server{
		Addr:         config.Global.ServerAddr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,   // Max time to read the request
		WriteTimeout: 0,                 // No timeout, allows long-lived SSE. Maybe use http.NewResponseController on SSE handlers
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}

	log.Info("server started", "addr", config.Global.ServerAddr)
	defer log.Info("server shutdown complete")

	eg.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-egctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Info("shutting down server...")

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("error during shutdown", "error", err)
			return err
		}

		return nil
	})

	return eg.Wait()
}
