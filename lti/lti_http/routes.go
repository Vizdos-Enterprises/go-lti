package lti_http

import (
	internal "github.com/vizdos-enterprises/go-lti/internal/adapters/server"

	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// WithProtectedRoutes registers LTI-protected routes (requires valid launch JWT).
func WithProtectedRoutes(routes ...lti_ports.ProtectedRoute) lti_ports.HTTPRouteOption {
	return internal.WithProtectedRoutes(routes...)
}
