package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_domain"
	"github.com/kvizdos/lti-server/lti/lti_ports"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
)

var (
	_ lti_ports.AsymetricSignerVerifier = (*KMSSigner)(nil)
)

// --- Minimal KMS interface for testability ---

type KMSClient interface {
	GetPublicKey(ctx context.Context, params *kms.GetPublicKeyInput, optFns ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(ctx context.Context, params *kms.SignInput, optFns ...func(*kms.Options)) (*kms.SignOutput, error)
}

type GetPublicKeyInput struct {
	KeyId *string
}

type GetPublicKeyOutput struct {
	PublicKey []byte
	KeyId     *string
}

type SignInput struct {
	KeyId            *string
	Message          []byte
	MessageType      string
	SigningAlgorithm string
}

type SignOutput struct {
	Signature []byte
}

// --- Constants for algorithms and message types ---

const (
	MessageTypeRaw              = "RAW"
	AlgECDSA_SHA256             = "ECDSA_SHA_256"
	AlgRSASSA_PKCS1_V1_5_SHA256 = "RSASSA_PKCS1_V1_5_SHA_256"
	DefaultSigningAlgorithm     = AlgECDSA_SHA256
)

// --- KMSSigner implementation ---

type KMSSigner struct {
	kmsConfig     *jwtkms.Config
	kmsClient     *kms.Client
	kmsKeyId      string
	issuer        string
	signingMethod *jwtkms.KMSSigningMethod
}

type KMSOption func(*KMSSigner)

func NewKMS(opts ...KMSOption) (*KMSSigner, error) {
	k := &KMSSigner{}
	for _, opt := range opts {
		opt(k)
	}
	if k.kmsConfig == nil {
		return nil, errors.New("kms config is required")
	}
	return k, nil
}

// --- Option helpers ---

// A configured kms client pointer to AWS KMS
// kmsClient KMSClient
// AWS KMS Key ID to be used
// kmsKeyID string
// If set to true JWT verification will be performed using KMS's Verify method
//
// In normal scenarios this can be left on the default false value, which will get, cache(forever) in memory and
// use the KMS key's public key to verify signatures
// verifyWithKMS bool
func WithKMS(kmsClient *kms.Client, kmsKeyID string, verifyWithKMS bool) KMSOption {
	return func(k *KMSSigner) {
		k.kmsClient = kmsClient
		k.kmsKeyId = kmsKeyID
		k.kmsConfig = jwtkms.NewKMSConfig(kmsClient, kmsKeyID, verifyWithKMS)
	}
}

func WithSigningMethod(method *jwtkms.KMSSigningMethod) KMSOption {
	return func(k *KMSSigner) { k.signingMethod = method }
}

func WithIssuer(issuer string) KMSOption {
	return func(k *KMSSigner) { k.issuer = issuer }
}

// --- Core methods ---

func (k *KMSSigner) getKID() string {
	kid := k.kmsKeyId
	if strings.HasPrefix(kid, "arn:aws:kms:") {
		parts := strings.Split(kid, "/")
		kid = parts[len(parts)-1]
	}
	return kid
}

func (k *KMSSigner) JWKs(ctx context.Context) (*lti_domain.JWKS, error) {
	pubOut, err := k.kmsClient.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: &k.kmsKeyId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	pub, err := x509.ParsePKIXPublicKey(pubOut.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	kid := k.getKID()

	jwk := lti_domain.JWK{
		Use: "sig",
		Kid: kid,
	}

	switch pk := pub.(type) {
	case *ecdsa.PublicKey:
		var crv, alg string
		switch pubOut.KeySpec {
		case "ECC_NIST_P256":
			crv, alg = "P-256", "ES256"
		case "ECC_NIST_P384":
			crv, alg = "P-384", "ES384"
		case "ECC_NIST_P521":
			crv, alg = "P-521", "ES512"
		default:
			return nil, fmt.Errorf("unsupported EC key spec: %v", pubOut.KeySpec)
		}

		xBytes := pk.X.FillBytes(make([]byte, (pk.Curve.Params().BitSize+7)/8))
		yBytes := pk.Y.FillBytes(make([]byte, (pk.Curve.Params().BitSize+7)/8))

		jwk.Kty = "EC"
		jwk.Crv = crv
		jwk.Alg = alg
		jwk.X = base64.RawURLEncoding.EncodeToString(xBytes)
		jwk.Y = base64.RawURLEncoding.EncodeToString(yBytes)

	case *rsa.PublicKey:
		nBytes := pk.N.Bytes()
		eBytes := big.NewInt(int64(pk.E)).Bytes()

		jwk.Kty = "RSA"
		jwk.Alg = "RS256" // most KMS RSA keys default to RSASSA_PKCS1_V1_5_SHA_256
		jwk.N = base64.RawURLEncoding.EncodeToString(nBytes)
		jwk.E = base64.RawURLEncoding.EncodeToString(eBytes)

	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}

	return &lti_domain.JWKS{Keys: []lti_domain.JWK{jwk}}, nil
}

func (k *KMSSigner) GetIssuer() string { return k.issuer }

func (k *KMSSigner) Sign(claims jwt.Claims, ttl time.Duration) (string, error) {
	if rc, ok := claims.(*jwt.RegisteredClaims); ok {
		if rc.Issuer == "" {
			rc.Issuer = k.issuer
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

	token := jwt.NewWithClaims(k.signingMethod, claims)

	token.Header["kid"] = k.getKID()

	signed, err := token.SignedString(k.kmsConfig.WithContext(context.Background()))
	if err != nil {
		return "", fmt.Errorf("kms sign failed: %w", err)
	}
	return signed, nil
}

func (k *KMSSigner) Verify(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != k.signingMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return k.kmsConfig, nil
	})
}
