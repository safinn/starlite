package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"starlite/internal/config"
	"starlite/internal/db"
	"starlite/internal/middleware"
	"starlite/internal/router"
	"starlite/pkg/logger"
	"starlite/pkg/nats"

	"github.com/alexedwards/scs/v2"
	natsgo "github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	toolbeltdb "github.com/delaneyj/toolbelt/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "app: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	l := logger.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
	ctx := context.Background()

	db, err := db.SetupDB(ctx, l.Logger, cfg.DataPath, cfg.IsDev())
	if err != nil {
		return fmt.Errorf("error setting up database: %w", err)
	}

	ns, err := nats.SetupNATS(ctx, l.Logger, cfg.DataPath)
	if err != nil {
		return err
	}

	// Create the single NATS client here so a connection failure aborts boot,
	// and features receive a ready *nats.Conn rather than the embedded server.
	nc, err := ns.Client()
	if err != nil {
		return fmt.Errorf("error creating nats client: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := serve(ctx, l.Logger, cfg, db, nc); err != nil &&
		!errors.Is(err, http.ErrServerClosed) &&
		!errors.Is(err, context.Canceled) {
		l.ErrorContext(ctx, "server failed", "err", err)
		return fmt.Errorf("serving: %w", err)
	}
	return nil
}

func serve(ctx context.Context, log *slog.Logger, cfg config.Config, db *toolbeltdb.Database, nc *natsgo.Conn) error {
	eg, egctx := errgroup.WithContext(ctx)
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour * 30
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = cfg.IsProd()

	mux := http.NewServeMux()

	if err := router.SetupRoutes(egctx, cfg, mux, sessionManager, db, nc, log); err != nil {
		return fmt.Errorf("error setting up routes: %w", err)
	}

	handler := middleware.Chain(mux,
		middleware.Recover(log),
		middleware.SecurityHeadersMiddleware(cfg.IsDev()),
		middleware.Logging(log),
		sessionManager.LoadAndSave,
	)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,   // Max time to read the request
		WriteTimeout: 0,                 // No timeout, allows long-lived SSE. Maybe use http.NewResponseController on SSE handlers
		IdleTimeout:  120 * time.Second, // Max time for keep-alive connections
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	log.InfoContext(ctx, "server ready", "addr", ln.Addr().String())

	eg.Go(func() error {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-egctx.Done()
		log.InfoContext(ctx, "server shutting down")

		if !cfg.IsDev() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			hard := make(chan os.Signal, 1)
			signal.Notify(hard, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-hard
				os.Exit(1)
			}()

			return srv.Shutdown(shutdownCtx)
		}

		return srv.Close()
	})

	return eg.Wait()
}
