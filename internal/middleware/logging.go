package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"starlite/pkg/logger"

	"github.com/felixge/httpsnoop"
)

func nextRequestID() string {
	var b [16]byte
	rand.Read(b[:])
	return fmt.Sprintf("%x", b[:])
}

// Logging returns middleware that logs each HTTP request with a request ID,
// status, duration, and any attributes accumulated on the request context
// via logctx.Set.
func Logging(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.NewLogCtx(r.Context())
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = nextRequestID()
			}
			w.Header().Set("X-Request-ID", reqID)
			logger.Set(ctx, slog.String("request_id", reqID))
			r = r.WithContext(ctx)

			start := time.Now()
			m := httpsnoop.CaptureMetrics(next, w, r)
			duration := time.Since(start)

			level, msg := classify(m.Code)
			l.LogAttrs(ctx, level, msg,
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", m.Code),
				slog.Duration("duration", duration),
				slog.Int64("bytes_written", m.Written),
			)
		})
	}
}

func classify(statusCode int) (slog.Level, string) {
	switch {
	case statusCode >= 500:
		return slog.LevelError, "http request error"
	case statusCode >= 400:
		return slog.LevelWarn, "http request client error"
	default:
		return slog.LevelInfo, "http request"
	}
}
