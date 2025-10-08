package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/server/middleware"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

// --- Fake verifier that can simulate good/bad responses ---

type fakeVerifier struct {
	shouldError   bool
	shouldBeValid bool
	audience      []string
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
	mw := middleware.VerifyLTI(v, []string{"tool.example"}, next)

	w := callMiddleware(t, mw)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 missing cookie, got %d", w.Code)
	}
	if called {
		t.Fatal("next handler should not have been called")
	}
}

func TestVerifyLTI_InvalidToken(t *testing.T) {
	v := &fakeVerifier{shouldError: true}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, next)
	cookie := &http.Cookie{Name: "lti_token", Value: "fake.jwt"}

	w := callMiddleware(t, mw, cookie)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 invalid token, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next handler should not run")
	}
}

func TestVerifyLTI_InvalidAudience(t *testing.T) {
	v := &fakeVerifier{shouldBeValid: true, audience: []string{"other.audience"}}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalled = true })

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, next)
	cookie := &http.Cookie{Name: "lti_token", Value: "valid.jwt"}

	w := callMiddleware(t, mw, cookie)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 invalid audience, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next handler should not run")
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

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, next)
	cookie := &http.Cookie{Name: "lti_token", Value: "good.jwt"}

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

	mw := middleware.VerifyLTI(v, []string{"tool.example"}, next)
	cookie := &http.Cookie{Name: "lti_token", Value: "bad.jwt"}

	w := callMiddleware(t, mw, cookie)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token validity, got %d", w.Code)
	}
	if called {
		t.Fatal("expected handler not to be invoked")
	}
}
