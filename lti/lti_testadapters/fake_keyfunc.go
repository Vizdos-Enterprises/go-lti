package lti_testadapters

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

// Ensure compile-time conformance.
var _ lti_ports.KeyfuncProvider = FakeKeyfuncProvider
var _ lti_ports.Keyfunc = (*fakeKeyfunc)(nil)

// fakeKeyfunc is a simple stub that always returns a fixed symmetric key.
type fakeKeyfunc struct{}

// Keyfunc satisfies the lti_ports.Keyfunc interface.
func (f *fakeKeyfunc) Keyfunc(_ *jwt.Token) (any, error) {
	return []byte("test-secret"), nil
}

// FakeKeyfuncProvider returns a deterministic Keyfunc implementation.
// It ignores the provided URLs and never makes network requests.
func FakeKeyfuncProvider(_ context.Context, _ []string) (lti_ports.Keyfunc, error) {
	return &fakeKeyfunc{}, nil
}
