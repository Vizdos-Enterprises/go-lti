package registry

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

// Ensure it implements *both* interfaces
var (
	_ lti_ports.Registry       = (*inMemoryRegistry)(nil)
	_ lti_ports.EphemeralStore = (*inMemoryRegistry)(nil)
)

// inMemoryRegistry implements Registry and EphemeralStore with thread-safe maps.
type inMemoryRegistry struct {
	mu sync.RWMutex

	deployments    map[string]lti_domain.Deployment // key: clientID|deploymentID
	state          map[string]stateRecord
	swapTokens     map[string]*lti_domain.SwapToken
	exchangeTokens map[string]*lti_domain.ExchangeToken
}

type stateRecord struct {
	data      lti_domain.State
	expiresAt time.Time
}

// NewInMemoryRegistry creates a new in-memory registry.
func NewInMemoryRegistry() lti_ports.EphemeralRegistry {
	return &inMemoryRegistry{
		deployments:    make(map[string]lti_domain.Deployment),
		state:          make(map[string]stateRecord),
		swapTokens:     make(map[string]*lti_domain.SwapToken),
		exchangeTokens: make(map[string]*lti_domain.ExchangeToken),
	}
}

func makeDeploymentKey(clientID, deploymentID string) string {
	return clientID + "|" + deploymentID
}

// ====================
//  Registry interface
// ====================

func (r *inMemoryRegistry) GetDeployment(ctx context.Context, clientID, deploymentID string) (lti_domain.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeDeploymentKey(clientID, deploymentID)
	dep, ok := r.deployments[key]
	if !ok {
		return nil, lti_domain.ErrDeploymentNotFound
	}
	return dep, nil
}

// ====================
//  EphemeralStore interface
// ====================

func (r *inMemoryRegistry) SaveSwapToken(ctx context.Context, swapToken string, data lti_domain.SwapToken, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.swapTokens[swapToken] = &data
	return nil
}

func (r *inMemoryRegistry) GetAndDeleteSwapToken(ctx context.Context, swapToken string) (*lti_domain.SwapToken, error) {
	if v, ok := r.swapTokens[swapToken]; ok {
		delete(r.swapTokens, swapToken)
		return v, nil
	}

	return nil, lti_domain.ErrSwapTokenNotFound
}

func (r *inMemoryRegistry) SaveExchangeToken(ctx context.Context, exchangeTokenID string, data lti_domain.ExchangeToken, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.exchangeTokens[exchangeTokenID] = &data
	return nil
}

func (r *inMemoryRegistry) ClaimExchangeToken(ctx context.Context, exchangeTokenID string, challenge string) (string, error) {
	exch, ok := r.exchangeTokens[exchangeTokenID]
	if !ok {
		return "", lti_domain.ErrExchangeTokenNotFound
	}

	if exch.Exchanged {
		return "", lti_domain.ErrExchangeTokenAlreadyExchanged
	}

	if exch.ClaimableUntil.Before(time.Now().UTC()) {
		return "", lti_domain.ErrExchangeRedemptionExpired
	}

	authToken := rand.Text()

	exch.AuthToken = authToken
	exch.Exchanged = true
	exch.Challenge = challenge
	r.mu.Lock()
	defer r.mu.Unlock()

	r.exchangeTokens[exchangeTokenID] = exch
	return authToken, nil
}

func (r *inMemoryRegistry) GetAndDeleteExchangeToken(ctx context.Context, exchangeTokenID string) (*lti_domain.ExchangeToken, error) {
	if v, ok := r.exchangeTokens[exchangeTokenID]; ok {
		delete(r.exchangeTokens, exchangeTokenID)
		return v, nil
	}

	return nil, lti_domain.ErrExchangeTokenNotFound
}

func (r *inMemoryRegistry) SaveState(ctx context.Context, stateID string, data lti_domain.State, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state[stateID] = stateRecord{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *inMemoryRegistry) DeleteState(ctx context.Context, stateID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.state[stateID]; !ok {
		return lti_domain.ErrStateNotFound
	}
	delete(r.state, stateID)
	return nil
}

func (r *inMemoryRegistry) GetState(ctx context.Context, stateID string) (*lti_domain.State, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.state[stateID]
	if !ok {
		return nil, lti_domain.ErrStateNotFound
	}
	if time.Now().After(rec.expiresAt) {
		return nil, lti_domain.ErrStateNotFound
	}
	return &rec.data, nil
}

// ====================
//  Helpers for seeding
// ====================

func (r *inMemoryRegistry) AddDeployment(ctx context.Context, dep lti_domain.Deployment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := makeDeploymentKey(dep.GetLTIClientID(), dep.GetLTIDeploymentID())
	r.deployments[key] = dep
}
