package crypto

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_ports"
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
func NewES256(keyID string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, issuer string) lti_ports.SignerVerifier {
	return &ES256Signer{
		privateKey: priv,
		publicKey:  pub,
		keyID:      keyID,
		issuer:     issuer,
	}
}

func (s ES256Signer) Asymetric() {}

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
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	})
	return token, err
}
