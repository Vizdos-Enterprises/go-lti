package crypto

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var (
	_ lti_ports.AsymetricSigner = (*ES256Signer)(nil)
)

// ES256Signer implements ports.SignerVerifier using an ECDSA (P-256) key.
type ES256Signer struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	keyID      string
	issuer     string
}

// NewES256 creates a new ES256 signer/verifier using an ECDSA keypair.
func NewES256(keyID string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, issuer string) lti_ports.AsymetricSignerVerifier {
	return &ES256Signer{
		privateKey: priv,
		publicKey:  pub,
		keyID:      keyID,
		issuer:     issuer,
	}
}

func (k *ES256Signer) JWKs(ctx context.Context) (*lti_domain.JWKS, error) {
	pub := k.publicKey
	if pub == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	// Convert big.Int coordinates to 32-byte base64url strings (no padding).
	xBytes := pub.X.FillBytes(make([]byte, 32))
	yBytes := pub.Y.FillBytes(make([]byte, 32))

	jwk := lti_domain.JWK{
		Kty: "EC",
		Crv: "P-256",
		Use: "sig",
		Alg: "ES256",
		Kid: k.keyID,
		X:   base64.RawURLEncoding.EncodeToString(xBytes),
		Y:   base64.RawURLEncoding.EncodeToString(yBytes),
	}

	return &lti_domain.JWKS{Keys: []lti_domain.JWK{jwk}}, nil
}

func (s *ES256Signer) GetIssuer() string {
	return s.issuer
}

// Sign generates an ES256-signed JWT from the provided claims.
func (s *ES256Signer) Sign(claims jwt.Claims, ttl time.Duration) (string, error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	return token.SignedString(s.privateKey)
}

// Verify parses and validates an ES256 token using the provided public key.
func (s *ES256Signer) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return s.publicKey, nil
	})
	return token, err
}
