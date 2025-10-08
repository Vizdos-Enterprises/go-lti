package lti_ports

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

// Signer defines an interface for generating and verifying internal JWTs.
type Signer interface {
	// Sign signs the given claims
	Sign(claims jwt.Claims, ttl time.Duration) (string, error)

	// Get the issuer assigned to this signer
	GetIssuer() string
}

type AsymetricSigner interface {
	Signer
	JWKs(context.Context) (*lti_domain.JWKS, error)
}

type Verifier interface {
	// Verify validates and parses a JWT, returning its claims.
	Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error)
}

type AsymetricVerifier interface {
	Verifier
	JWKs(context.Context) (*lti_domain.JWKS, error)
}

type SignerVerifier interface {
	Signer
	Verifier
}

type AsymetricSignerVerifier interface {
	AsymetricSigner
	AsymetricVerifier
}
