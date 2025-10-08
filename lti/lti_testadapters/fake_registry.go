package lti_testadapters

import (
	"context"
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
	States      sync.Map
	Deployments sync.Map
}

func (f *FakeRegistry) GetDeployment(_ context.Context, clientID, depID string) (*lti_domain.Deployment, error) {
	v, ok := f.Deployments.Load(depID)
	if !ok {
		return nil, lti_domain.ErrDeploymentNotFound
	}
	dep := v.(lti_domain.Deployment)
	return &dep, nil
}

func (f *FakeRegistry) AddDeployment(ctx context.Context, dep *lti_domain.Deployment) {
	f.Deployments.Store(dep.InternalID, dep)
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

// Test Helpers
func (f *FakeRegistry) AddDeploymentQuick(clientID, deploymentID, issuer, jwksURL, tenantID string) {
	f.Deployments.Store(deploymentID, lti_domain.Deployment{
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
