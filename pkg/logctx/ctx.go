package logctx

import (
	"context"
	"log/slog"
	"sync"
)

type ctxKey struct{}

type Attrs struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

func (a *Attrs) Add(attrs ...slog.Attr) {
	a.mu.Lock()
	a.attrs = append(a.attrs, attrs...)
	a.mu.Unlock()
}

func (a *Attrs) All() []slog.Attr {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.attrs
}

func New(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &Attrs{})
}

func Set(ctx context.Context, attrs ...slog.Attr) {
	if a, ok := ctx.Value(ctxKey{}).(*Attrs); ok {
		a.Add(attrs...)
	}
}

func Get(ctx context.Context) []slog.Attr {
	if a, ok := ctx.Value(ctxKey{}).(*Attrs); ok {
		return a.All()
	}
	return nil
}
