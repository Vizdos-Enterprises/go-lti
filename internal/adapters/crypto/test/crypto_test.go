package crypto_test

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

// runSignerVerifierTests runs a shared suite against any lti_ports.SignerVerifier.
func runSignerVerifierTests(t *testing.T, name string, s lti_ports.SignerVerifier) {
	t.Run(name+"/SignAndVerify", func(t *testing.T) {
		claims := &jwt.RegisteredClaims{
			Subject: "student42",
		}

		tokenStr, err := s.Sign(claims, time.Minute)
		if err != nil {
			t.Fatalf("sign failed: %v", err)
		}
		if tokenStr == "" {
			t.Fatal("expected signed token string, got empty")
		}
		if !strings.Contains(tokenStr, ".") {
			t.Fatalf("expected JWT structure with '.', got %q", tokenStr)
		}

		parsedClaims := &jwt.RegisteredClaims{}
		token, err := s.Verify(tokenStr, parsedClaims)
		if err != nil {
			t.Fatalf("verify failed: %v", err)
		}
		if !token.Valid {
			t.Fatal("expected token to be valid")
		}
		if parsedClaims.Issuer == "" {
			t.Fatal("expected issuer to be populated")
		}
	})

	t.Run(name+"/RejectsWrongAlg", func(t *testing.T) {
		// Craft an invalid algorithm header.
		header := `{"alg":"none","typ":"JWT"}`
		payload := `{"sub":"123"}`
		tokenStr := fmt.Sprintf("%s.%s.",
			base64.RawURLEncoding.EncodeToString([]byte(header)),
			base64.RawURLEncoding.EncodeToString([]byte(payload)),
		)

		_, err := s.Verify(tokenStr, &jwt.RegisteredClaims{})
		if err == nil {
			t.Fatalf("expected error verifying alg=none token")
		}
	})

	t.Run(name+"/SetsDefaults", func(t *testing.T) {
		claims := &jwt.RegisteredClaims{}
		tokenStr, err := s.Sign(claims, 2*time.Second)
		if err != nil {
			t.Fatal(err)
		}

		var parsed jwt.RegisteredClaims
		token, err := s.Verify(tokenStr, &parsed)
		if err != nil {
			t.Fatal(err)
		}
		if !token.Valid {
			t.Fatal("expected token valid")
		}
		if parsed.Issuer == "" {
			t.Error("expected default issuer to be set")
		}
		if parsed.IssuedAt == nil || parsed.ExpiresAt == nil {
			t.Error("expected default iat/exp to be set")
		}
	})
}
