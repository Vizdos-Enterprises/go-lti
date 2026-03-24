package telemetry

import "github.com/vizdos-enterprises/go-lti/lti/lti_domain"

type NoopTelemetry struct{}

func (t NoopTelemetry) EmitLaunch(event lti_domain.LaunchEvent) {
	// noop
}

func (t NoopTelemetry) Events() <-chan lti_domain.LaunchEvent {
	return nil
}
