package lti_telemetry

import (
	"github.com/vizdos-enterprises/go-lti/internal/adapters/telemetry"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func NewAsyncTelemetry(bufferSize int) lti_ports.TelemetryPort {
	return telemetry.NewLaunchEmitter(bufferSize)
}
