package server

import "github.com/kvizdos/lti-server/lti/lti_ports"

type ServerOption func(*Server)

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		launcher: nil, // must be set via option
	}

	for _, opt := range opts {
		opt(s)
	}

	// sanity check: launcher is required
	if s.launcher == nil {
		panic("lti: launcher must be provided (use WithLauncher)")
	}

	if s.verifier == nil {
		panic("lti: verifier must be provided (use WithVerifier)")
	}

	return s
}

func WithLauncher(l lti_ports.Launcher) ServerOption {
	return func(s *Server) {
		s.launcher = l
	}
}

func WithVerifier(ver lti_ports.AsymetricVerifier) ServerOption {
	return func(s *Server) {
		s.verifier = ver
	}
}
