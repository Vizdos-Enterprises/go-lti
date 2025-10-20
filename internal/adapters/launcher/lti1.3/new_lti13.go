package launcher1dot3

import (
	"fmt"
	"strings"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/keyfunc"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/redirector"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/registry"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_logger"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

type LauncherOptions func(*LTI13_Launcher)

func NewLauncher(opts ...LauncherOptions) *LTI13_Launcher {
	l := &LTI13_Launcher{
		redirector: redirector.NewDefaultRedirector(""),
		enabledServices: []lti_domain.LTIService{
			lti_domain.LTIService_ResourceLink,
		},
	}
	for _, opt := range opts {
		opt(l)
	}

	if l.baseURL == "" {
		panic("baseURL is required for a launcher. Call WithBaseURL")
	}

	if l.logger == nil {
		l.logger = lti_logger.NewNoopLogger()
	}

	if l.registry == nil {
		l.registry = registry.NewInMemoryRegistry()
	}

	if l.ephemeral == nil {
		l.ephemeral = registry.NewInMemoryRegistry()
	}

	if l.redirector == nil {
		l.redirector = redirector.NewDefaultRedirector(fmt.Sprintf("%s/lti/app", strings.TrimRight(l.baseURL, "/")))
	}

	if l.signer == nil {
		panic("a signer is required for a launcher. Call with WithSigner")
	}

	if l.keyfunc == nil {
		l.keyfunc = keyfunc.DefaultKeyfuncProviderAdapter()
	}

	return l
}

func WithBaseURL(baseURL string) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.baseURL = baseURL
	}
}

func WithLogger(logger lti_ports.Logger) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.logger = logger
	}
}

func WithRegistry(registry lti_ports.Registry) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.registry = registry
	}
}

func WithEphemeralStorage(ephemeral lti_ports.EphemeralStore) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.ephemeral = ephemeral
	}
}

func WithRedirector(redirector lti_ports.Redirector) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.redirector = redirector
	}
}

func WithRedirectURL(redirectURL string) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.redirector = redirector.NewDefaultRedirector(redirectURL)
	}
}

func WithSigner(signer lti_ports.Signer) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.signer = signer
	}
}

func WithAudience(audience []string) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.audience = audience
	}
}

func WithKeyFunc(keyfuncProvider lti_ports.KeyfuncProvider) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.keyfunc = keyfuncProvider
	}
}

func WithImpostering(imposterJWT *lti_domain.LTIJWT) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.imposterJWT = imposterJWT
	}
}

func WithDeepLinking(deepLinkingService lti_ports.DeepLinking) LauncherOptions {
	return func(s *LTI13_Launcher) {
		s.enabledServices = append(s.enabledServices, lti_domain.LTIService_DeepLink)
		s.deepLinkingService = deepLinkingService
	}
}
