package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/server"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_http"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// --- Fakes --------------------------------------------------------------

type fakeLauncher struct {
	versionCalled bool
	launchCalled  bool
	oidcCalled    bool
}

func (f *fakeLauncher) GetLTIVersion() string {
	f.versionCalled = true
	return "1.3"
}
func (f *fakeLauncher) GetAudience() []string { return []string{"aud"} }
func (f *fakeLauncher) HandleLaunch(w http.ResponseWriter, _ *http.Request) {
	f.launchCalled = true
	w.WriteHeader(http.StatusOK)
}
func (f *fakeLauncher) HandleOIDC(w http.ResponseWriter, _ *http.Request) {
	f.oidcCalled = true
	w.WriteHeader(http.StatusOK)
}

type fakeVerifier struct{}

func (f *fakeVerifier) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	// Return a minimal valid token
	return &jwt.Token{
		Valid:  true,
		Method: jwt.SigningMethodHS256,
	}, nil
}

func (f *fakeVerifier) JWKs(_ context.Context) (*lti_domain.JWKS, error) {
	return nil, nil
}

// --- Tests --------------------------------------------------------------

func TestCreateRoutes_BasicRoutes(t *testing.T) {
	launcher := &fakeLauncher{}
	verifier := &fakeVerifier{}
	s := server.NewServer(server.WithLauncher(launcher), server.WithVerifier(verifier))

	mux := s.CreateRoutes()
	tests := []struct {
		path       string
		wantStatus int
		wantCalled *bool
	}{
		{"/lti/1.3/launch", http.StatusOK, &launcher.launchCalled},
		{"/lti/1.3/oidc", http.StatusOK, &launcher.oidcCalled},
		{"/lti/nonexistent", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.wantCalled != nil && !*tt.wantCalled {
				t.Fatalf("expected handler for %s to be called", tt.path)
			}
		})
	}
}

func TestCreateRoutes_WithProtectedRoutes_CustomVerifier(t *testing.T) {
	launcher := &fakeLauncher{}
	verifier := &fakeVerifier{}
	s := server.NewServer(server.WithLauncher(launcher), server.WithVerifier(verifier))

	called := false

	customVerifier := func(_ lti_ports.Verifier, _ []string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			// Inject a fake LTI session
			fakeSession := &lti_domain.LTIJWT{
				Roles: []lti_domain.Role{lti_domain.MEMBERSHIP_LEARNER},
			}
			ctx := lti_domain.ContextWithLTI(r.Context(), fakeSession)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	mux := s.CreateRoutes(lti_http.WithProtectedRoutes(
		lti_ports.ProtectedRoute{
			Path:     "/test",
			Role:     nil, // no role restriction for this test
			Handler:  handler,
			Verifier: customVerifier,
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/lti/app/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !called {
		t.Fatalf("expected custom verifier to be used")
	}
	if w.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 from protected route, got %d", w.Result().StatusCode)
	}
}
