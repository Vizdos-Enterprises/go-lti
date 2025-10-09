package lti_http

import (
	internal "github.com/vizdos-enterprises/go-lti/internal/adapters/server"

	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

type opt struct {
	toInternal func() lti_ports.HTTPRouteOptions
}

func (o opt) ToInternal() lti_ports.HTTPRouteOptions {
	return o.toInternal()
}

// WithProtectedRoutes registers LTI-protected routes (requires valid launch JWT).
func WithProtectedRoutes(routes ...lti_ports.ProtectedRoute) lti_ports.HTTPRouteOption {
	return opt{toInternal: func() lti_ports.HTTPRouteOptions {
		return internal.WithProtectedRoutes(routes...)
	}}
}
