package lti_testadapters

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// Ensure interface conformance
var _ lti_ports.SignerVerifier = (*FakeSigner)(nil)

// FakeSigner is a minimal stub for lti_ports.SignerVerifier.
// It always returns deterministic results for predictable tests.
type FakeSigner struct {
	Issuer            string
	LastSigned        jwt.Claims
	ShouldError       bool // if true, Sign/Verify will return an error
	ReturnSignedValue string
}

func (f *FakeSigner) GetIssuer() string {
	if f.Issuer == "" {
		return "tool"
	}
	return f.Issuer
}

// Sign returns a dummy JWT string and captures the claims for assertions.
func (f *FakeSigner) Sign(c jwt.Claims, _ time.Duration) (string, error) {
	if f.ShouldError {
		return "", fmt.Errorf("fake signer error")
	}
	f.LastSigned = c
	return f.ReturnSignedValue, nil
}

// Verify pretends to verify the token and can simulate success or failure.
func (f *FakeSigner) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	if f.ShouldError {
		return nil, fmt.Errorf("fake verifier error")
	}

	switch dst := claims.(type) {
	case *jwt.RegisteredClaims:
		dst.Issuer = f.GetIssuer()
		dst.Subject = "fake-user"
		dst.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
	}

	// Return a minimally valid token object.
	token := &jwt.Token{
		Valid:  true,
		Method: jwt.SigningMethodHS256,
	}
	return token, nil
}

// MustHaveSigned is a test helper to assert that Sign() was called.
func (f *FakeSigner) MustHaveSigned(t *testing.T) {
	t.Helper()
	if f.LastSigned == nil {
		t.Fatalf("expected FakeSigner.Sign() to be called but it was not")
	}
}
