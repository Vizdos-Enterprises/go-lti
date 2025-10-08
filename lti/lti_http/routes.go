package lti_http

import (
	"net/http"

	internal "github.com/kvizdos/lti-server/internal/adapters/server"

	"github.com/kvizdos/lti-server/lti/lti_ports"
)

// HTTPRouteOption represents a route configuration option.
type HTTPRouteOption struct {
	toInternal func() internal.HTTPRouteOptions
}

// WithProtectedRoutes registers LTI-protected routes (requires valid launch JWT).
func WithProtectedRoutes(handler http.Handler, customVerifierFunc ...lti_ports.VerifyTokenFunc) HTTPRouteOption {
	return HTTPRouteOption{toInternal: func() internal.HTTPRouteOptions {
		return internal.WithProtectedRoutes(handler, customVerifierFunc...)
	}}
}
