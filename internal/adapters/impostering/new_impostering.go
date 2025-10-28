package impostering

import (
	"github.com/vizdos-enterprises/go-lti/lti/lti_logger"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewImpostering(opts ...lti_ports.ImposteringOption) lti_ports.Impostering {
	l := &ImposteringService{
		logger: lti_logger.NewNoopLogger(),
	}
	for _, opt := range opts {
		opt(l)
	}

	if l.sessionSigner == nil {
		panic("a session signer is required for a launcher. Call with WithSessionSigner")
	}

	if l.incomingVerifier == nil {
		panic("an incoming verifier is required for a launcher. Call with WithIncomingVerifier")
	}

	return l
}

func WithIncomingVerifier(verifier lti_ports.Verifier) lti_ports.ImposteringOption {
	return func(l lti_ports.Impostering) {
		cast := l.(*ImposteringService)
		cast.incomingVerifier = verifier
	}
}

func WithSessionSigner(signer lti_ports.Signer) lti_ports.ImposteringOption {
	return func(l lti_ports.Impostering) {
		cast := l.(*ImposteringService)
		cast.sessionSigner = signer
	}
}

func WithIncomingAudience(audience []string) lti_ports.ImposteringOption {
	return func(l lti_ports.Impostering) {
		cast := l.(*ImposteringService)
		cast.audience = audience
	}
}

func WithSessionAudience(audience []string) lti_ports.ImposteringOption {
	return func(l lti_ports.Impostering) {
		cast := l.(*ImposteringService)
		cast.sessionAud = audience
	}
}

func WithLogger(logger lti_ports.Logger) lti_ports.ImposteringOption {
	return func(l lti_ports.Impostering) {
		cast := l.(*ImposteringService)
		cast.logger = logger
	}
}
