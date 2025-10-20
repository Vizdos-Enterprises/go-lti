package lti_deeplink

import (
	"github.com/vizdos-enterprises/go-lti/internal/adapters/deeplinking"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewDeepLinkingService(opts ...lti_ports.DeepLinkingOption) lti_ports.DeepLinking {
	return deeplinking.NewDeepLinkingService(opts...)
}

// WithSigner sets the signer used to generate internal JWTs.
func WithSigner(signer lti_ports.AsymetricSigner) lti_ports.DeepLinkingOption {
	return deeplinking.WithSigner(signer)
}

func WithRedirectURL(url string) lti_ports.DeepLinkingOption {
	return deeplinking.WithRedirectURL(url)
}
