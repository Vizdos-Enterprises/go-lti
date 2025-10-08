package lti_crypto

import (
	"github.com/kvizdos/lti-server/internal/adapters/crypto"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

func NewHMAC(keyID string, secret string, issuer string) lti_ports.SignerVerifier {
	return crypto.NewHMAC(keyID, secret, issuer)
}
