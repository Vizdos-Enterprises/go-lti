package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/kvizdos/lti-server/internal/adapters/crypto"
)

func TestCryptoImplementation(t *testing.T) {
	hmac := crypto.NewHMAC("key123", "supersecret", "issuer.example")
	runSignerVerifierTests(t, "HMAC", hmac)

	// RS256 implementation
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	rs256 := crypto.NewRS256("rsa-key", priv, &priv.PublicKey, "issuer.example")
	runSignerVerifierTests(t, "RS256", rs256)
}
