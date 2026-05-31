// Package logger provides context-scoped log attribute accumulation for use with
// slog. Attributes added to a context via Set are automatically attached to
// log records emitted with InfoContext / ErrorContext / etc., allowing
// middleware and deep-stack code to enrich logs without explicit logger
// propagation.
//
// Typical use:
//
//	// In main, wrap your handler:
//	base := slog.NewJSONHandler(os.Stderr, opts)
//	l := slog.New(log.NewContextHandler(base))
//
//	// At request entry:
//	ctx = log.NewLogCtx(ctx)
//
//	// Anywhere downstream:
//	logger.Set(ctx, slog.String("user_id", uid))
//
//	// When logging — attributes are picked up automatically:
//	l.InfoContext(ctx, "request done")
package logger

import (
	"context"
	"log/slog"
	"sync"
)

// ctxKey is the unexported type used as the context key for storing Attrs.
// Using an unexported struct type prevents collisions with keys from other
// packages, since no one else can construct a ctxKey{}.
type ctxKey struct{}

// Attrs is a mutable, goroutine-safe collection of slog attributes that travels
// in a context.Context. Handlers (e.g. middleware, request scopes) append to it
// as a request flows through the system, and the ContextHandler reads them
// when emitting log records.
type Attrs struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

// Add appends one or more attributes to the collection. Safe to call from
// multiple goroutines concurrently.
func (a *Attrs) Add(attrs ...slog.Attr) {
	a.mu.Lock()
	a.attrs = append(a.attrs, attrs...)
	a.mu.Unlock()
}

// All returns a copy of the accumulated attributes. The returned slice is
// owned by the caller and is safe to mutate or retain.
func (a *Attrs) All() []slog.Attr {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]slog.Attr, len(a.attrs))
	copy(out, a.attrs)
	return out
}

// NewLogCtx returns a new context carrying an empty Attrs collection. Call this
// once at the start of a logical operation (e.g. an incoming HTTP request) so
// that downstream code can attach attributes via Set that will be picked up
// by the ContextHandler when log records are emitted on the same context tree.
func NewLogCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &Attrs{})
}

// Set adds attributes to the Attrs collection stored on ctx, if any. It is a
// no-op when ctx was not created with NewLogCtx. Safe to call from multiple
// goroutines concurrently.
func Set(ctx context.Context, attrs ...slog.Attr) {
	if a, ok := ctx.Value(ctxKey{}).(*Attrs); ok {
		a.Add(attrs...)
	}
}

// Get returns a copy of the accumulated attributes from ctx, or nil if ctx
// was not created with NewLogCtx.
func Get(ctx context.Context) []slog.Attr {
	if a, ok := ctx.Value(ctxKey{}).(*Attrs); ok {
		return a.All()
	}
	return nil
}
