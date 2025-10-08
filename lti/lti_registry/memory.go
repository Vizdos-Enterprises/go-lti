package lti_registry

import (
	"github.com/vizdos-enterprises/go-lti/internal/adapters/registry"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewMemoryRegistry() lti_ports.EphemeralRegistry {
	return registry.NewInMemoryRegistry()
}
