package logger

import (
	"context"
	"log/slog"
)

// ContextHandler is a slog.Handler that enriches each log record with
// attributes accumulated on the context via Set. It wraps an inner handler
// and delegates all real work to it.
type ContextHandler struct {
	inner slog.Handler
}

// NewContextHandler returns a ContextHandler wrapping inner. Use it when
// constructing your slog.Logger:
//
//	logger := slog.New(log.NewContextHandler(slog.NewJSONHandler(os.Stderr, opts)))
func NewContextHandler(inner slog.Handler) *ContextHandler {
	return &ContextHandler{inner: inner}
}

// Enabled reports whether the handler handles records at the given level.
// It delegates to the inner handler.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle adds any context-scoped attributes from ctx to the record and then
// forwards it to the inner handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := Get(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.inner.Handle(ctx, r)
}

// WithAttrs returns a new ContextHandler whose inner handler has the given
// attributes pre-attached. Used by slog when you call logger.With(...).
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new ContextHandler whose inner handler has the given
// group name applied. Used by slog when you call logger.WithGroup(...).
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{inner: h.inner.WithGroup(name)}
}
