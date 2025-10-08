package lti_ports

import (
	"net/http"
)

type HTTPRouteOptions func(mux *http.ServeMux)

type Server interface {
	CreateRoutes(opts ...HTTPRouteOptions) *http.ServeMux
}

type VerifyTokenFunc func(verifier Verifier, expectedAudience []string, next http.Handler) http.Handler
