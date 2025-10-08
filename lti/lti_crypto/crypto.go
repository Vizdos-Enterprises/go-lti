package lti_crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/crypto"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewHMAC(keyID string, secret string, issuer string) lti_ports.SignerVerifier {
	return crypto.NewHMAC(keyID, secret, issuer)
}

func NewRS256(keyID string, priv *rsa.PrivateKey, pub *rsa.PublicKey, issuer string) lti_ports.AsymetricSignerVerifier {
	return crypto.NewRS256(keyID, priv, pub, issuer)
}

func NewES256(keyID string, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, issuer string) lti_ports.AsymetricSignerVerifier {
	return crypto.NewES256(keyID, priv, pub, issuer)
}
