package lti_launcher

import (
	launcher1dot3 "github.com/vizdos-enterprises/go-lti/internal/adapters/launcher/lti1.3"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// Launcher represents any LTI-compliant launcher implementation.
type Launcher = lti_ports.Launcher

// NewLTI13Launcher returns a new LTI 1.3-compliant launcher configured with the provided options.
// It wraps the internal implementation and exposes a stable public API.
func NewLTI13Launcher(opts ...LauncherOption) Launcher {
	internalOpts := []launcher1dot3.LauncherOptions{}

	for _, opt := range opts {
		internalOpts = append(internalOpts, opt.toInternal())
	}

	return launcher1dot3.NewLauncher(internalOpts...)
}

// LauncherOption represents a configurable option for creating a new LTI launcher.
type LauncherOption struct {
	toInternal func() launcher1dot3.LauncherOptions
}

// WithBaseURL sets the base URL for LTI login and launch redirects.
func WithBaseURL(baseURL string) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithBaseURL(baseURL)
	}}
}

// WithLogger sets the logger implementation.
func WithLogger(logger lti_ports.Logger) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithLogger(logger)
	}}
}

// WithRegistry sets the registry used for deployments and tenants.
func WithRegistry(reg lti_ports.Registry) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithRegistry(reg)
	}}
}

// WithEphemeralStorage sets the store used for transient states and nonces.
func WithEphemeralStorage(store lti_ports.EphemeralStore) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithEphemeralStorage(store)
	}}
}

// WithRedirectURL defines where the tool should redirect after a successful LTI launch.
func WithRedirectURL(redirectURL string) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithRedirectURL(redirectURL)
	}}
}

// WithSigner sets the signer used to generate internal JWTs.
func WithSigner(signer lti_ports.Signer) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithSigner(signer)
	}}
}

// WithAudience sets the audience (aud claim) for internal JWTs.
func WithAudience(audience []string) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithAudience(audience)
	}}
}

// WithImpostering sets whether impostering is enabled.
func WithImpostering(imposteringJWT *lti_domain.LTIJWT) LauncherOption {
	return LauncherOption{toInternal: func() launcher1dot3.LauncherOptions {
		return launcher1dot3.WithImpostering(imposteringJWT)
	}}
}
