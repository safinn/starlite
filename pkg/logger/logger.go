package logger

import (
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

type Logger struct {
	*slog.Logger
	level *slog.LevelVar
}

func (l *Logger) SetLevel(s string) {
	l.level.Set(parseLogLevel(s))
}

func New(w io.Writer, level string, format string) *Logger {
	levelVar := new(slog.LevelVar)
	levelVar.Set(parseLogLevel(level))

	var base slog.Handler
	switch format {
	case "json":
		base = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: levelVar,
		})
	default:
		base = tint.NewHandler(w, &tint.Options{
			Level:      levelVar,
			TimeFormat: time.Kitchen,
		})
	}

	return &Logger{
		Logger: slog.New(NewContextHandler(base)),
		level:  levelVar,
	}
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
