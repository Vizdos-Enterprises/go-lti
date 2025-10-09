package lti_ports

import (
	"net/http"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

type HTTPRouteOption func(server Server, mux *http.ServeMux)

type Server interface {
	CreateRoutes(opts ...HTTPRouteOption) *http.ServeMux

	GetLauncher() Launcher
	GetVerifier() Verifier
}

type VerifyTokenFunc func(verifier Verifier, expectedAudience []string, next http.Handler) http.Handler

type ProtectedRoute struct {
	Path     string
	Role     []lti_domain.Role
	Handler  http.Handler
	Verifier VerifyTokenFunc
}
