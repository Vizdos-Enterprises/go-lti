package telemetry

import (
	"sync/atomic"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

type LaunchEmitter struct {
	ch      chan lti_domain.LaunchEvent
	dropped atomic.Uint64
}

func NewLaunchEmitter(buffer int) *LaunchEmitter {
	return &LaunchEmitter{
		ch: make(chan lti_domain.LaunchEvent, buffer),
	}
}

func (e *LaunchEmitter) EmitLaunch(ev lti_domain.LaunchEvent) {
	select {
	case e.ch <- ev:
	default:
		e.dropped.Add(1)
	}
}

func (e *LaunchEmitter) Events() <-chan lti_domain.LaunchEvent {
	return e.ch
}

func (e *LaunchEmitter) Dropped() uint64 {
	return e.dropped.Load()
}
