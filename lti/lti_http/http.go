package lti_http

import (
	"net/http"

	internal "github.com/vizdos-enterprises/go-lti/internal/adapters/server"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var _ lti_ports.Server = (*HTTPServer)(nil)

// HTTPServer represents a running LTI HTTP server.
// It wraps the internal implementation and provides a stable public API.
type HTTPServer struct {
	inner *internal.Server
}

func (h HTTPServer) GetLauncher() lti_ports.Launcher {
	return h.inner.GetLauncher()
}

func (h HTTPServer) GetVerifier() lti_ports.Verifier {
	return h.inner.GetVerifier()
}

// NewServer constructs a new LTI Server using the provided options.
// It panics if required fields (launcher, verifier) are missing.
func NewServer(opts ...ServerOption) *HTTPServer {
	internalOpts := []internal.ServerOption{}
	for _, o := range opts {
		internalOpts = append(internalOpts, o.toInternal())
	}
	return &HTTPServer{inner: internal.NewServer(internalOpts...)}
}

// CreateRoutes builds a ServeMux with all LTI endpoints and any protected routes.
func (s *HTTPServer) CreateRoutes(opts ...lti_ports.HTTPRouteOption) *http.ServeMux {
	return s.inner.CreateRoutes(opts...)
}
