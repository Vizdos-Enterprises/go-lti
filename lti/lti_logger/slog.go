package lti_logger

import (
	"log/slog"
	"os"

	"github.com/kvizdos/lti-server/lti/lti_ports"
)

type slogLogger struct {
	l *slog.Logger
}

// Compile-time check that we satisfy ports.Logger
var _ lti_ports.Logger = (*slogLogger)(nil)

type SlogLoggerOption func(*slogLogger)

// Constructor
func NewSlogLogger(opts ...SlogLoggerOption) lti_ports.Logger {
	// Use default text handler writing to stderr
	handler := slog.NewTextHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)

	o := &slogLogger{
		l: slog.New(handler),
	}

	// Apply options
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Implement interface methods
func (s *slogLogger) Info(msg string, kv ...any)  { s.l.Info(msg, kv...) }
func (s *slogLogger) Warn(msg string, kv ...any)  { s.l.Warn(msg, kv...) }
func (s *slogLogger) Debug(msg string, kv ...any) { s.l.Debug(msg, kv...) }
func (s *slogLogger) Error(msg string, kv ...any) { s.l.Error(msg, kv...) }

func WithSlogHandler(handler slog.Handler) SlogLoggerOption {
	return func(s *slogLogger) {
		s.l = slog.New(handler)
	}
}
