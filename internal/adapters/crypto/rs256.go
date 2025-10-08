package crypto

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_domain"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

var (
	_ lti_ports.SignerVerifier          = (*RS256Signer)(nil)
	_ lti_ports.AsymetricSignerVerifier = (*RS256Signer)(nil)
)

// RS256Signer implements lti_ports.SignerVerifier using RSA (RS256).
type RS256Signer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	issuer     string
}

// NewRS256 creates a new RSA signer/verifier pair.
func NewRS256(keyID string, priv *rsa.PrivateKey, pub *rsa.PublicKey, issuer string) lti_ports.SignerVerifier {
	return &RS256Signer{
		privateKey: priv,
		publicKey:  pub,
		keyID:      keyID,
		issuer:     issuer,
	}
}

func (s *RS256Signer) JWKs(ctx context.Context) (*lti_domain.JWKS, error) {
	pub := s.publicKey
	if pub == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	// Encode modulus (N) and exponent (E) in base64url without padding.
	nBytes := pub.N.Bytes()
	eBytes := big.NewInt(int64(pub.E)).Bytes()

	jwk := lti_domain.JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: s.keyID,
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}

	return &lti_domain.JWKS{Keys: []lti_domain.JWK{jwk}}, nil
}

func (s *RS256Signer) GetIssuer() string {
	return s.issuer
}

// Sign creates a JWT using RS256 and the provided claims.
func (s *RS256Signer) Sign(claims jwt.Claims, ttl time.Duration) (string, error) {
	// Apply sensible defaults for registered claims.
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
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	return token.SignedString(s.privateKey)
}

// Verify parses and validates an RS256 token using the public key.
func (s *RS256Signer) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return s.publicKey, nil
	})
	return token, err
}
