package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/server/middleware"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

type Server struct {
	launcher    lti_ports.Launcher
	verifier    lti_ports.AsymetricVerifier
	impostering lti_ports.Impostering
	mux         http.ServeMux
}

func (s *Server) CreateRoutes(opts ...lti_ports.HTTPRouteOption) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/lti/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	if s.impostering != nil {
		mux.HandleFunc("/lti/imposter", s.impostering.HandleImposterLaunch)
	}

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

	if s.verifier != nil {
		mux.HandleFunc("/lti/keys.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			keys, err := s.verifier.JWKs(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			js, err := json.Marshal(keys)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(js)
		})
	}

	return mux
}

func (s *Server) GetVerifier() lti_ports.Verifier {
	return s.verifier
}

func (s *Server) GetLauncher() lti_ports.Launcher {
	return s.launcher
}

func WithProtectedRoutes(routes ...lti_ports.ProtectedRoute) lti_ports.HTTPRouteOption {
	return func(s lti_ports.Server, m *http.ServeMux) {
		for _, route := range routes {
			var vFunc lti_ports.VerifyTokenFunc
			vFunc = middleware.VerifyLTI
			if route.Verifier != nil {
				vFunc = route.Verifier
			}

			// First wrap the handler with RequireRole
			roleChecked := middleware.RequireRole(route.Role...)(route.Handler)

			// Then wrap the result with the verifier
			protected := vFunc(s.GetVerifier(), s.GetLauncher().GetAudience(), route.AllowImpostering, roleChecked)
			path := fmt.Sprintf("/lti/app%s", route.Path)
			strip := fmt.Sprintf("/lti/app%s", strings.TrimRight(route.Path, "/"))
			m.Handle(path, http.StripPrefix(strip, protected))
		}
	}
}
