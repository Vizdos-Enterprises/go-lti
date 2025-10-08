package lti_registry

import (
	"github.com/kvizdos/lti-server/internal/adapters/registry"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

func NewMemoryRegistry() lti_ports.EphemeralRegistry {
	return registry.NewInMemoryRegistry()
}
