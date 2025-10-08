package lti_http

import (
	internal "github.com/vizdos-enterprises/go-lti/internal/adapters/server"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// ServerOption represents a public configuration option for the LTI Server.
type ServerOption struct {
	toInternal func() internal.ServerOption
}

// WithLauncher sets the launcher implementation.
func WithLauncher(l lti_ports.Launcher) ServerOption {
	return ServerOption{toInternal: func() internal.ServerOption {
		return internal.WithLauncher(l)
	}}
}

// WithVerifier sets the token verifier implementation.
func WithVerifier(v lti_ports.AsymetricVerifier) ServerOption {
	return ServerOption{toInternal: func() internal.ServerOption {
		return internal.WithVerifier(v)
	}}
}
