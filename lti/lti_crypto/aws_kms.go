package lti_crypto

import (
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/crypto"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// KMSOptions represents a configurable option for creating a new KMS.
type KMSOption struct {
	toInternal func() crypto.KMSOption
}

func NewKMS(opts ...KMSOption) (lti_ports.AsymetricSignerVerifier, error) {
	internalOpts := []crypto.KMSOption{}

	for _, opt := range opts {
		internalOpts = append(internalOpts, opt.toInternal())
	}

	return crypto.NewKMS(internalOpts...)
}

func WithKMS(kmsClient *kms.Client, kmsKeyID string, verifyWithKMS bool) KMSOption {
	return KMSOption{toInternal: func() crypto.KMSOption {
		return crypto.WithKMS(kmsClient, kmsKeyID, verifyWithKMS)
	}}
}

func WithSigningMethod(method *jwtkms.KMSSigningMethod) KMSOption {
	return KMSOption{toInternal: func() crypto.KMSOption {
		return crypto.WithSigningMethod(method)
	}}
}

func WithIssuer(issuer string) KMSOption {
	return KMSOption{toInternal: func() crypto.KMSOption {
		return crypto.WithIssuer(issuer)
	}}
}
