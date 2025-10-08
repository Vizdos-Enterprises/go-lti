package crypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/crypto"
)

func TestCryptoImplementation(t *testing.T) {
	hmac := crypto.NewHMAC("key123", "supersecret", "issuer.example")
	runSignerVerifierTests(t, "HMAC", hmac)

	// RS256 implementation
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	rs256 := crypto.NewRS256("rsa-key", priv, &priv.PublicKey, "issuer.example")
	runSignerVerifierTests(t, "RS256", rs256)

	eccpriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	eccpub := &eccpriv.PublicKey

	es256 := crypto.NewES256("ecc-key-1", eccpriv, eccpub, "https://tool.example")
	runSignerVerifierTests(t, "ES256", es256)
}
