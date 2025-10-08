package launcher1dot3_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	launcher1dot3 "github.com/vizdos-enterprises/go-lti/internal/adapters/launcher/lti1.3"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_testadapters"
)

// setupLauncher creates a configured launcher with all fake adapters.
func setupLauncher() (*launcher1dot3.LTI13_Launcher, *lti_testadapters.FakeRegistry, *lti_testadapters.FakeRedirect, *lti_testadapters.FakeSigner, *lti_testadapters.FakeLogger) {
	reg := &lti_testadapters.FakeRegistry{}
	redir := &lti_testadapters.FakeRedirect{}
	signer := &lti_testadapters.FakeSigner{ReturnSignedValue: "signed.jwt"}
	logger := lti_testadapters.NewFakeLogger()

	reg.AddDeploymentQuick("client1", "dep1", "https://lms.example", "https://jwks.example", "tenantA")

	l := launcher1dot3.NewLauncher(
		launcher1dot3.WithBaseURL("https://tool.example"),
		launcher1dot3.WithRegistry(reg),
		launcher1dot3.WithEphemeralStorage(reg),
		launcher1dot3.WithRedirector(redir),
		launcher1dot3.WithSigner(signer),
		launcher1dot3.WithLogger(logger),
		launcher1dot3.WithKeyFunc(lti_testadapters.FakeKeyfuncProvider),
	)

	return l, reg, redir, signer, logger
}

func TestHandleOIDC_Success(t *testing.T) {
	l, reg, _, _, _ := setupLauncher()

	form := url.Values{
		"iss":               {"https://lms.example"},
		"client_id":         {"client1"},
		"lti_deployment_id": {"dep1"},
		"login_hint":        {"hint"},
		"target_link_uri":   {"https://tool.example/lti/launch"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oidc", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	l.HandleOIDC(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, "response_type=id_token") {
		t.Errorf("expected redirect with OIDC params, got %s", loc)
	}

	if reg.CountStates() != 1 {
		t.Errorf("expected 1 saved state, got %d", reg.CountStates())
	}
}

func TestHandleOIDC_InvalidIssuer(t *testing.T) {
	l, _, _, _, logger := setupLauncher()

	form := url.Values{
		"iss":               {"https://wrong.example"},
		"client_id":         {"client1"},
		"lti_deployment_id": {"dep1"},
		"login_hint":        {"hint"},
		"target_link_uri":   {"https://tool.example/lti/launch"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oidc", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	l.HandleOIDC(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid issuer, got %d", w.Result().StatusCode)
	}
	if !logger.ContainsMessage("Invalid issuer") {
		t.Errorf("expected 'Invalid issuer' log, got %+v", logger.Entries())
	}
}

func TestHandleLaunch_Success(t *testing.T) {
	l, reg, redir, signer, logger := setupLauncher()

	// Pre-store valid state
	stateID := reg.AddStateQuick("", lti_domain.State{
		Issuer:       "https://lms.example",
		ClientID:     "client1",
		DeploymentID: "dep1",
		Nonce:        "nonce-123",
		TenantID:     "tenantA",
		CreatedAt:    time.Now(),
	})

	// Build fake JWT with valid nonce & claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "user123",
		"nonce": "nonce-123",
		"https://purl.imsglobal.org/spec/lti/claim/message_type": "LtiResourceLinkRequest",
		"https://purl.imsglobal.org/spec/lti/claim/context": map[string]any{
			"id": "course1", "label": "C101", "title": "Intro to Testing",
		},
		"https://purl.imsglobal.org/spec/lti/claim/roles": []any{"Instructor"},
	})
	rawToken, _ := token.SignedString([]byte("test-secret"))

	form := url.Values{"id_token": {rawToken}, "state": {stateID}}
	req := httptest.NewRequest(http.MethodPost, "/launch", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	l.HandleLaunch(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		t.Logf("%+v\n", logger.Entries())
		t.Fatalf("expected redirect or 200 OK, got %d", resp.StatusCode)
	}
	if !redir.DidRedirect() {
		t.Fatalf("expected RedirectAfterLaunch to be called")
	}
	if !redir.HasToken("signed.jwt") {
		t.Fatalf("expected redirect to include signed.jwt")
	}
	signer.MustHaveSigned(t)
}

func TestHandleLaunch_MissingParams(t *testing.T) {
	l, _, _, _, _ := setupLauncher()

	req := httptest.NewRequest(http.MethodPost, "/launch", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	l.HandleLaunch(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 missing params, got %d", w.Result().StatusCode)
	}
}

func TestHandleLaunch_InvalidState(t *testing.T) {
	l, _, _, _, logger := setupLauncher()

	form := url.Values{"id_token": {"some.token"}, "state": {"does-not-exist"}}
	req := httptest.NewRequest(http.MethodPost, "/launch", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	l.HandleLaunch(w, req)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 invalid state, got %d", w.Result().StatusCode)
	}
	if !logger.ContainsMessage("Invalid or expired state") {
		t.Errorf("expected 'Invalid or expired state' log entry")
	}
}
