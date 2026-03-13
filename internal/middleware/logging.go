// Package middleware provides HTTP middleware components for request processing.
package middleware

import (
	"log/slog"
	"net/http"
	"starlite/pkg/logctx"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// Flush implements the http.Flusher interface to support SSE
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
		// We don't call rw.ResponseWriter.WriteHeader(http.StatusOK) here
		// because Write() automatically sends a 200 OK if WriteHeader hasn't been called.
	}
	return rw.ResponseWriter.Write(b)
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}

// LoggingMiddleware returns an HTTP middleware that logs requests using the provided logger.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logctx.New(r.Context())
			r = r.WithContext(ctx)

			start := time.Now()

			// Wrap the response writer to capture status code
			rw := newResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Calculate request duration
			duration := time.Since(start)

			level, msg := generateLevelAndMessage(rw.statusCode)
			logger.LogAttrs(ctx, level, msg,
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Int64("duration_ms", duration.Milliseconds()),
				// slog.String("remote_addr", r.Header.Get("X-Real-IP")),
				// slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

func generateLevelAndMessage(statusCode int) (slog.Level, string) {
	msg := "http request"
	level := slog.LevelInfo

	if statusCode >= 500 {
		level = slog.LevelError
	} else if statusCode >= 400 {
		level = slog.LevelWarn
	}

	return level, msg
}
