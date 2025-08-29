package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// Logger wraps slog with custom configuration
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(level slog.Level) *Logger {
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: "02/01/2006 15:04:05",
	})

	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

// SetDefault sets the logger as the default slog logger
func (l *Logger) SetDefault() {
	slog.SetDefault(l.Logger)
}

// ParseLevel parses string level to slog.Level
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
