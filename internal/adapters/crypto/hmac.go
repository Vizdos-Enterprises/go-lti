package crypto

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

var (
	_ lti_ports.SignerVerifier = (*HMACSigner)(nil)
)

// HMACSigner implements ports.Signer using an HMAC secret key.
type HMACSigner struct {
	key    []byte
	keyID  string
	issuer string
}

// NewHMAC creates a new symmetric HMAC signer.
func NewHMAC(keyID, secret, issuer string) lti_ports.SignerVerifier {
	return &HMACSigner{
		key:    []byte(secret),
		keyID:  keyID,
		issuer: issuer,
	}
}

func (s *HMACSigner) GetIssuer() string {
	return s.issuer
}

// Sign generates a signed JWT from the provided claims.
// Typically used with *domain.LTIJWT.
func (s *HMACSigner) Sign(claims jwt.Claims, ttl time.Duration) (string, error) {
	// Enforce exp/iat defaults if the struct has none
	if rc, ok := claims.(*jwt.RegisteredClaims); ok {
		if rc.Issuer == "" {
			rc.Issuer = s.issuer
		}
		if rc.IssuedAt == nil {
			rc.IssuedAt = jwt.NewNumericDate(time.Now())
		}
		if rc.ExpiresAt == nil && ttl.Seconds() > 0 {
			rc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(ttl))
		}
		if rc.NotBefore == nil {
			rc.NotBefore = jwt.NewNumericDate(time.Now())
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	return token.SignedString(s.key)
}

// Verify parses and validates a token against this signer’s key.
// You must pass an empty claims struct (e.g., &domain.LTIJWT{}).
func (s *HMACSigner) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return s.key, nil
	})
	return token, err
}
