package lti_http

import (
	internal "github.com/kvizdos/lti-server/internal/adapters/server"

	"github.com/kvizdos/lti-server/lti/lti_ports"
)

// HTTPRouteOption represents a route configuration option.
type HTTPRouteOption struct {
	toInternal func() internal.HTTPRouteOptions
}

// WithProtectedRoutes registers LTI-protected routes (requires valid launch JWT).
func WithProtectedRoutes(routes ...lti_ports.ProtectedRoute) HTTPRouteOption {
	return HTTPRouteOption{toInternal: func() internal.HTTPRouteOptions {
		return internal.WithProtectedRoutes(routes...)
	}}
}
