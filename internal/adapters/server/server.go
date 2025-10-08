package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kvizdos/lti-server/internal/adapters/server/middleware"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

type Server struct {
	launcher lti_ports.Launcher
	verifier lti_ports.AsymetricVerifier
	mux      http.ServeMux
}

type HTTPRouteOptions func(server *Server, mux *http.ServeMux)

func (s *Server) CreateRoutes(opts ...HTTPRouteOptions) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/lti/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	mux.HandleFunc(fmt.Sprintf("/lti/%s/launch", s.launcher.GetLTIVersion()), s.launcher.HandleLaunch)
	mux.HandleFunc(fmt.Sprintf("/lti/%s/oidc", s.launcher.GetLTIVersion()), s.launcher.HandleOIDC)
	mux.HandleFunc("/lti/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jwks, err := s.verifier.JWKs(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		js, err := json.Marshal(jwks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	})

	for _, opt := range opts {
		opt(s, mux)
	}

	return mux
}

func WithProtectedRoutes(routes ...lti_ports.ProtectedRoute) HTTPRouteOptions {
	return func(s *Server, m *http.ServeMux) {
		for _, route := range routes {
			var vFunc lti_ports.VerifyTokenFunc
			vFunc = middleware.VerifyLTI
			if route.Verifier != nil {
				vFunc = route.Verifier
			}

			// First wrap the handler with RequireRole
			roleChecked := middleware.RequireRole(route.Role...)(route.Handler)

			// Then wrap the result with the verifier
			protected := vFunc(s.verifier, s.launcher.GetAudience(), roleChecked)
			path := fmt.Sprintf("/lti/app%s", route.Path)
			strip := fmt.Sprintf("/lti/app%s", strings.TrimRight(route.Path, "/"))
			m.Handle(path, http.StripPrefix(strip, protected))
		}
	}
}
