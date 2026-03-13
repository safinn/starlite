// Package logger provides application-wide structured logging configuration.
package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"starlite/internal/config"
	"starlite/pkg/logctx"

	"github.com/lmittmann/tint"
)

var logLevel slog.LevelVar // zero value is Info

// New creates and returns a new slog logger based on the provided configuration.
// In production, logs are formatted as JSON with Info level.
// In all other environments, logs are formatted as text for better readability.
func New() *slog.Logger {
	var handler slog.Handler
	SetLogLevel(getLogLevel(config.Global.LogLevel))

	if config.Global.Env == config.Prod {
		// JSON format for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: &logLevel,
		})
	} else {
		// Text format for development
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      &logLevel,
			TimeFormat: time.Kitchen,
		})
	}

	logger := slog.New(logctx.NewHandler(handler))

	return logger
}

func getLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func SetLogLevel(level slog.Level) {
	logLevel.Set(level)
}
