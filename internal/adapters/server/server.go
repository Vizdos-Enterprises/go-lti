package server

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func WithProtectedRoutes(handler http.Handler, customVerifierFunc ...lti_ports.VerifyTokenFunc) HTTPRouteOptions {
	return func(s *Server, m *http.ServeMux) {
		var vFunc lti_ports.VerifyTokenFunc
		vFunc = middleware.VerifyLTI
		if len(customVerifierFunc) > 0 && customVerifierFunc[0] != nil {
			vFunc = customVerifierFunc[0]
		}
		protected := vFunc(s.verifier, s.launcher.GetAudience(), handler)
		m.Handle("/lti/app/", http.StripPrefix("/lti/app", protected))
	}
}
