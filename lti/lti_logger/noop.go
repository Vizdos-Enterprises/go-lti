package lti_logger

import (
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

type noopLogger struct{}

// Compile-time check that we satisfy ports.Logger
var _ lti_ports.Logger = (*noopLogger)(nil)

// Constructor
func NewNoopLogger() lti_ports.Logger {
	return &noopLogger{}
}

// Implement interface methods
func (s *noopLogger) Info(msg string, kv ...any)  {}
func (s *noopLogger) Warn(msg string, kv ...any)  {}
func (s *noopLogger) Debug(msg string, kv ...any) {}
func (s *noopLogger) Error(msg string, kv ...any) {}
