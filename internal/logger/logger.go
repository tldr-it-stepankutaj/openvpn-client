package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

// Logger wraps slog.Logger with additional context
type Logger struct {
	*slog.Logger
}

// Options for configuring the logger
type Options struct {
	Level   slog.Level
	JSON    bool
	Output  io.Writer
	Program string
}

// New creates a new logger with the given options
func New(opts Options) *Logger {
	var handler slog.Handler

	handlerOpts := &slog.HandlerOptions{
		Level: opts.Level,
	}

	if opts.Output == nil {
		opts.Output = os.Stderr
	}

	if opts.JSON {
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	} else {
		handler = slog.NewTextHandler(opts.Output, handlerOpts)
	}

	// Add program name as default attribute
	if opts.Program != "" {
		handler = handler.WithAttrs([]slog.Attr{
			slog.String("program", opts.Program),
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext returns a new context with the logger attached
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// With returns a logger with additional attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithUser returns a logger with user context
func (l *Logger) WithUser(username string) *Logger {
	return l.With("username", username)
}

// WithSession returns a logger with session context
func (l *Logger) WithSession(sessionID string) *Logger {
	return l.With("session_id", sessionID)
}

// WithError returns a logger with error context
func (l *Logger) WithError(err error) *Logger {
	return l.With("error", err.Error())
}

// Fatal logs at the error level and exits with code 1
func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs at the error level with context and exits with code 1
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}
