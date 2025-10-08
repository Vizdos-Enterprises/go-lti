package lti_crypto

import (
	"crypto/rsa"

	"github.com/kvizdos/lti-server/internal/adapters/crypto"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

func NewHMAC(keyID string, secret string, issuer string) lti_ports.SignerVerifier {
	return crypto.NewHMAC(keyID, secret, issuer)
}

func NewRS256(keyID string, priv *rsa.PrivateKey, pub *rsa.PublicKey, issuer string) lti_ports.SignerVerifier {
	return crypto.NewRS256(keyID, priv, pub, issuer)
}
