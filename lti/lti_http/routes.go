package lti_http

import (
	"github.com/vizdos-enterprises/go-lti/internal/adapters/helper_routes"
	internal "github.com/vizdos-enterprises/go-lti/internal/adapters/server"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// WithProtectedRoutes registers LTI-protected routes (requires valid launch JWT).
func WithProtectedRoutes(
	routes ...lti_ports.ProtectedRoute,
) lti_ports.HTTPRouteOption {
	return internal.WithProtectedRoutes(routes...)
}

func RegisterSessionInfoJS(distinctIdGenerator func(*lti_domain.LTIJWT) string) lti_ports.ProtectedRoute {
	return lti_ports.ProtectedRoute{
		Path:                   "/session-info.js",
		Role:                   []lti_domain.Role{},
		RequireDeepLinkContext: false,
		Handler:                helper_routes.NewSessionInitializerHTTP(distinctIdGenerator),
		Verifier:               nil,
		AllowImpostering:       true,
	}
}
