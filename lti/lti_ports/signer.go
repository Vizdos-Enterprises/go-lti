package lti_ports

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Signer defines an interface for generating and verifying internal JWTs.
type Signer interface {
	// Sign signs the given claims
	Sign(claims jwt.Claims, ttl time.Duration) (string, error)

	// Get the issuer assigned to this signer
	GetIssuer() string
}

type AssymetricSigner interface {
	Signer
	Assymetric()
}

type Verifier interface {
	// Verify validates and parses a JWT, returning its claims.
	Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error)
}

type AssymetricVerifier interface {
	Verifier
	Assymetric()
}

type SignerVerifier interface {
	Signer
	Verifier
}

type AssymetricSignerVerifier interface {
	AssymetricSigner
	AssymetricVerifier
}
