package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/server/middleware"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// --- Fake verifier that can simulate good/bad responses ---

var _ lti_ports.AsymetricSignerVerifier = (*fakeVerifier)(nil)

type fakeVerifier struct {
	shouldError   bool
	shouldBeValid bool
	audience      []string
}

func (fakeVerifier) Sign(claims jwt.Claims, ttl time.Duration) (string, error) {
	return "", nil
}

func (fakeVerifier) GetIssuer() string {
	return "tool.example"
}

func (f *fakeVerifier) JWKs(_ context.Context) (*lti_domain.JWKS, error) {
	return nil, nil
}

func (f *fakeVerifier) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	if f.shouldError {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// populate some claims
	if lti, ok := claims.(*lti_domain.LTIJWT); ok {
		lti.Audience = f.audience
	}

	tok := &jwt.Token{
		Valid: f.shouldBeValid,
		Claims: jwt.RegisteredClaims{
			Audience: f.audience,
		},
		Method: jwt.SigningMethodHS256,
	}
	return tok, nil
}

// --- Helper for executing the middleware ---

func callMiddleware(t *testing.T, mw http.Handler, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	return w
}

// --- Tests ---

func TestVerifyLTI_MissingCookie(t *testing.T) {
	v := &fakeVerifier{}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	mw := middleware.VerifyLTI(v, []string{"tool.example"}, true, next)

	w := callMiddleware(t, mw)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected StatusTemporaryRedirect for invalid token validity, got %d", w.Code)
	}

	if called {
		t.Fatal("expected handler not to be invoked")
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected redirect location header to be set")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}

	if u.Path != "/lti/auth/error" {
		t.Fatalf("expected redirect path %q, got %q", "/lti/auth/error", u.Path)
	}

	q := u.Query()
	if got := q.Get("err"); got != "missing token" {
		t.Fatalf("expected err=%q, got %q", "missing token", got)
	}
}

func TestVerifyLTI_InvalidToken(t *testing.T) {
	v := &fakeVerifier{shouldError: true}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, true, next)
	cookie := &http.Cookie{Name: lti_domain.ContextKey_Session, Value: "fake.jwt"}

	w := callMiddleware(t, mw, cookie)

	if nextCalled {
		t.Fatal("next handler should not run")
	}

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected StatusTemporaryRedirect for invalid token validity, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected redirect location header to be set")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}

	if u.Path != "/lti/auth/error" {
		t.Fatalf("expected redirect path %q, got %q", "/lti/auth/error", u.Path)
	}

	q := u.Query()
	if got := q.Get("err"); got != "invalid token" {
		t.Fatalf("expected err=%q, got %q", "invalid token", got)
	}
}

func TestVerifyLTI_InvalidAudience(t *testing.T) {
	v := &fakeVerifier{shouldBeValid: true, audience: []string{"other.audience"}}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, true, next)
	cookie := &http.Cookie{Name: lti_domain.ContextKey_Session, Value: "valid.jwt"}

	w := callMiddleware(t, mw, cookie)

	if nextCalled {
		t.Fatal("next handler should not run")
	}

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected StatusTemporaryRedirect for invalid token validity, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected redirect location header to be set")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}

	if u.Path != "/lti/auth/error" {
		t.Fatalf("expected redirect path %q, got %q", "/lti/auth/error", u.Path)
	}

	q := u.Query()
	if got := q.Get("err"); got != "could not verify audience" {
		t.Fatalf("expected err=%q, got %q", "could not verify audience", got)
	}
}

func TestVerifyLTI_Success(t *testing.T) {
	v := &fakeVerifier{shouldBeValid: true, audience: []string{"tool.example"}}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Ensure context has LTI claims
		if ctxClaims, ok := lti_domain.LTIFromContext(r.Context()); !ok || ctxClaims == nil {
			t.Fatal("expected LTI claims in context")
		}
	})

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, true, next)
	cookie := &http.Cookie{Name: lti_domain.ContextKey_Session, Value: "good.jwt"}

	w := callMiddleware(t, mw, cookie)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected next handler to be invoked")
	}
}

func TestVerifyLTI_TokenNotValid(t *testing.T) {
	v := &fakeVerifier{shouldBeValid: false}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, true, next)
	cookie := &http.Cookie{Name: lti_domain.ContextKey_Session, Value: "bad.jwt"}

	w := callMiddleware(t, mw, cookie)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected StatusTemporaryRedirect for invalid token validity, got %d", w.Code)
	}

	if called {
		t.Fatal("expected handler not to be invoked")
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected redirect location header to be set")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}

	if u.Path != "/lti/auth/error" {
		t.Fatalf("expected redirect path %q, got %q", "/lti/auth/error", u.Path)
	}

	q := u.Query()
	if got := q.Get("err"); got != "invalid token" {
		t.Fatalf("expected err=%q, got %q", "invalid token", got)
	}
}
