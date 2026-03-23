package lti_testadapters

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var (
	_ lti_ports.Registry       = (*FakeRegistry)(nil)
	_ lti_ports.EphemeralStore = (*FakeRegistry)(nil)
)

type FakeRegistry struct {
	States         sync.Map
	Deployments    sync.Map
	Swaps          sync.Map
	ExchangeTokens sync.Map

	lastSavedExchangeTokenID string
}

func (f *FakeRegistry) GetLastSavedExchangeTokenID() string {
	return f.lastSavedExchangeTokenID
}

func (f *FakeRegistry) GetDeployment(_ context.Context, clientID, depID string) (lti_domain.Deployment, error) {
	v, ok := f.Deployments.Load(depID)
	if !ok {
		return nil, lti_domain.ErrDeploymentNotFound
	}
	dep := v.(lti_domain.Deployment)
	return dep, nil
}

func (f *FakeRegistry) AddDeployment(ctx context.Context, dep lti_domain.Deployment) {
	f.Deployments.Store(dep.GetDeploymentID(), dep)
}

func (f *FakeRegistry) SaveState(_ context.Context, key string, value lti_domain.State, _ time.Duration) error {
	f.States.Store(key, value)
	return nil
}
func (f *FakeRegistry) GetState(_ context.Context, key string) (*lti_domain.State, error) {
	val, ok := f.States.Load(key)
	if !ok {
		return nil, lti_domain.ErrStateNotFound
	}
	v := val.(lti_domain.State)
	return &v, nil
}
func (f *FakeRegistry) DeleteState(_ context.Context, key string) error {
	f.States.Delete(key)
	return nil
}

func (f *FakeRegistry) SaveSwapToken(_ context.Context, swapToken string, data lti_domain.SwapToken, _ time.Duration) error {
	f.Swaps.Store(swapToken, &data)
	return nil
}
func (f *FakeRegistry) GetAndDeleteSwapToken(_ context.Context, swapToken string) (*lti_domain.SwapToken, error) {
	val, ok := f.Swaps.Load(swapToken)
	if !ok {
		return nil, lti_domain.ErrSwapTokenNotFound
	}
	f.Swaps.Delete(swapToken)
	return val.(*lti_domain.SwapToken), nil
}

func (f *FakeRegistry) SaveExchangeToken(ctx context.Context, exchangeTokenID string, data lti_domain.ExchangeToken, ttl time.Duration) error {
	f.lastSavedExchangeTokenID = exchangeTokenID
	f.ExchangeTokens.Store(exchangeTokenID, &data)
	return nil
}

func (r *FakeRegistry) ClaimExchangeToken(ctx context.Context, exchangeTokenID string, challenge string) (string, error) {
	v, ok := r.ExchangeTokens.Load(exchangeTokenID)
	if !ok {
		return "", lti_domain.ErrExchangeTokenNotFound
	}
	exch := v.(*lti_domain.ExchangeToken)
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

	r.ExchangeTokens.Store(exchangeTokenID, exch)
	return authToken, nil
}

func (r *FakeRegistry) GetAndDeleteExchangeToken(ctx context.Context, exchangeTokenID string) (*lti_domain.ExchangeToken, error) {
	fmt.Println("Loading", exchangeTokenID)
	v, ok := r.ExchangeTokens.LoadAndDelete(exchangeTokenID)
	if ok {
		return v.(*lti_domain.ExchangeToken), nil
	}
	return nil, lti_domain.ErrExchangeTokenNotFound
}

// Test Helpers
func (f *FakeRegistry) AddDeploymentQuick(clientID, deploymentID, issuer, jwksURL, tenantID string) {
	f.Deployments.Store(deploymentID, lti_domain.BaseLTIDeployment{
		ClientID:      clientID,
		DeploymentID:  deploymentID,
		Issuer:        issuer,
		JWKSURL:       jwksURL,
		AuthEndpoint:  fmt.Sprintf("%s/authorize", issuer),
		TokenEndpoint: fmt.Sprintf("%s/token", issuer),
		ForTenantID:   tenantID,
		InternalID:    deploymentID,
	})
}

func (f *FakeRegistry) AddStateQuick(key string, state lti_domain.State) string {
	if key == "" {
		key = fmt.Sprintf("state-%d", time.Now().UnixNano())
	}
	f.States.Store(key, state)
	return key
}

// MustGetStateT retrieves a state and fails the test if not found.
func (f *FakeRegistry) MustGetStateT(t *testing.T, key string) lti_domain.State {
	t.Helper()
	val, ok := f.States.Load(key)
	if !ok {
		t.Fatalf("state %q not found in FakeRegistry", key)
	}
	return val.(lti_domain.State)
}

// MustGetDeploymentT retrieves a deployment and fails the test if not found.
func (f *FakeRegistry) MustGetDeploymentT(t *testing.T, depID string) lti_domain.Deployment {
	t.Helper()
	val, ok := f.Deployments.Load(depID)
	if !ok {
		t.Fatalf("deployment %q not found in FakeRegistry", depID)
	}
	return val.(lti_domain.Deployment)
}

// Reset clears all stored deployments and states — ideal for test setup/teardown.
func (f *FakeRegistry) Reset() {
	f.States = sync.Map{}
	f.Deployments = sync.Map{}
}

// CountStates returns how many state entries exist.
func (f *FakeRegistry) CountStates() int {
	count := 0
	f.States.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// CountDeployments returns how many deployment entries exist.
func (f *FakeRegistry) CountDeployments() int {
	count := 0
	f.Deployments.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
