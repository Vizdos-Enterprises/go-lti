package lti_testadapters

import (
	"net/http"
	"testing"

	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var _ lti_ports.FallbackAuthorizer = (*FakeFallbackAuthorizer)(nil)

type FakeFallbackAuthorizer struct {
	FallbackCalled bool
	FallbackToken  string
}

func (f *FakeFallbackAuthorizer) MustHaveBeenCalled(t *testing.T) {
	if !f.FallbackCalled {
		t.Fatalf("expected fallback to be called, got false")
	}
}

func (f *FakeFallbackAuthorizer) HandleFallback(w http.ResponseWriter, r *http.Request, exchangeToken string) {
	f.FallbackCalled = true
	f.FallbackToken = exchangeToken
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte("fallback handler"))
}

func (f *FakeFallbackAuthorizer) Route() *http.ServeMux {
	return nil
}
