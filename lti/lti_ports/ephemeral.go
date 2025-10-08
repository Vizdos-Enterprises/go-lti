package lti_ports

import (
	"context"
	"time"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

type EphemeralStore interface {
	SaveState(ctx context.Context, stateID string, data lti_domain.State, ttl time.Duration) error
	DeleteState(ctx context.Context, stateID string) error
	GetState(ctx context.Context, stateID string) (*lti_domain.State, error)
}

type EphemeralRegistry interface {
	EphemeralStore
	Registry
}
