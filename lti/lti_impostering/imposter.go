package lti_impostering

import (
	"github.com/vizdos-enterprises/go-lti/internal/adapters/impostering"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewImpostering(options ...lti_ports.ImposteringOption) lti_ports.Impostering {
	return impostering.NewImpostering(options...)
}

func WithSessionSigner(signer lti_ports.Signer) lti_ports.ImposteringOption {
	return impostering.WithSessionSigner(signer)
}

func WithIncomingVerifier(verifier lti_ports.Verifier) lti_ports.ImposteringOption {
	return impostering.WithIncomingVerifier(verifier)
}

func WithIncomingAudience(audience []string) lti_ports.ImposteringOption {
	return impostering.WithIncomingAudience(audience)
}

func WithSessionAudience(audience []string) lti_ports.ImposteringOption {
	return impostering.WithSessionAudience(audience)
}

func WithLogger(logger lti_ports.Logger) lti_ports.ImposteringOption {
	return impostering.WithLogger(logger)
}
