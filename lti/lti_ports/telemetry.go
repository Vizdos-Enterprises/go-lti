package lti_ports

import "github.com/vizdos-enterprises/go-lti/lti/lti_domain"

type TelemetryPort interface {
	EmitLaunch(lti_domain.LaunchEvent)
	Events() <-chan lti_domain.LaunchEvent
}
