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

	SaveSwapToken(ctx context.Context, swapToken string, data lti_domain.SwapToken, ttl time.Duration) error
	GetAndDeleteSwapToken(ctx context.Context, swapToken string) (*lti_domain.SwapToken, error)

	SaveExchangeToken(ctx context.Context, exchangeToken string, data lti_domain.ExchangeToken, ttl time.Duration) error
	ClaimExchangeToken(ctx context.Context, exchangeTokenID string, challenge string) (authToken string, err error)
	GetAndDeleteExchangeToken(ctx context.Context, exchangeTokenID string) (*lti_domain.ExchangeToken, error)
}

type EphemeralRegistry interface {
	EphemeralStore
	Registry
}
