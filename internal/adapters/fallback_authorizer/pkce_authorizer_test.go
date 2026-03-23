package fallback_authorizer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_testadapters"
)

func setupAuthorizer() (*pkceAuthorizer, *lti_testadapters.FakeRegistry, *lti_testadapters.FakeSigner, *lti_testadapters.FakeLogger) {
	reg := &lti_testadapters.FakeRegistry{}
	signer := &lti_testadapters.FakeSigner{ReturnSignedValue: "signed.jwt"}
	logger := lti_testadapters.NewFakeLogger()

	p := New(reg, signer, logger)
	return p, reg, signer, logger
}

func mustJSONBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	return bytes.NewReader(b)
}

func validVerifierAndChallenge() (string, string) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~abc"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}

func decodeJSONMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	var out map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response json: %v; body=%s", err, rr.Body.String())
	}

	return out
}

func TestHandleFallback_RedirectsToVerify(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodGet, "/lti/auth/fallback", nil)
	w := httptest.NewRecorder()

	p.HandleFallback(w, req, "exchange-123")
	resp := w.Result()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected %d, got %d", http.StatusFound, resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if loc != "/lti/auth/verify?exchange=exchange-123" {
		t.Fatalf("expected redirect location %q, got %q", "/lti/auth/verify?exchange=exchange-123", loc)
	}
}

func TestInitExchangeCode_MethodNotAllowed(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodGet, "/init", nil)
	w := httptest.NewRecorder()

	p.initExchangeCode(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrMethodNotAllowed) {
		t.Fatalf("expected err %q, got %q", ErrMethodNotAllowed, body["err"])
	}
}

func TestInitExchangeCode_BadJSON(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/init", bytes.NewBufferString("{"))
	w := httptest.NewRecorder()

	p.initExchangeCode(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrBadRequest) {
		t.Fatalf("expected err %q, got %q", ErrBadRequest, body["err"])
	}
}

func TestInitExchangeCode_MissingParams(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/init", mustJSONBody(t, initRequest{
		ExchangeToken: "",
		Challenge:     "",
	}))
	w := httptest.NewRecorder()

	p.initExchangeCode(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrMissingParams) {
		t.Fatalf("expected err %q, got %q", ErrMissingParams, body["err"])
	}
}

func TestInitExchangeCode_ExchangeTokenNotFound(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/init", mustJSONBody(t, initRequest{
		ExchangeToken: "missing",
		Challenge:     "challenge",
	}))
	w := httptest.NewRecorder()

	p.initExchangeCode(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrFailedToExchange) {
		t.Fatalf("expected err %q, got %q", ErrFailedToExchange, body["err"])
	}
}

func TestInitExchangeCode_Success(t *testing.T) {
	p, reg, _, _ := setupAuthorizer()

	tokenID := "exchange-123"
	err := reg.SaveExchangeToken(context.Background(), tokenID, lti_domain.ExchangeToken{
		Data:           &lti_domain.SwapToken{},
		ClaimableUntil: time.Now().Add(5 * time.Minute),
		Exchanged:      false,
		Challenge:      "challenge-123",
		AuthToken:      "auth-token",
	}, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/init", mustJSONBody(t, initRequest{
		ExchangeToken: tokenID,
		Challenge:     "challenge-123",
	}))
	w := httptest.NewRecorder()

	p.initExchangeCode(w, req)

	if w.Result().StatusCode != http.StatusOK {
		resp := w.Body.String()
		t.Fatalf("expected 200, got %d, response: %q", w.Result().StatusCode, resp)
	}

	body := decodeJSONMap(t, w)
	if body["token"] == "" {
		t.Fatalf("expected auth token in response, got empty string")
	}
}

func TestExchangeForToken_MethodNotAllowed(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodGet, "/exchange", nil)
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrMethodNotAllowed) {
		t.Fatalf("expected err %q, got %q", ErrMethodNotAllowed, body["err"])
	}
}

func TestExchangeForToken_BadJSON(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/exchange", bytes.NewBufferString("{"))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrBadRequest) {
		t.Fatalf("expected err %q, got %q", ErrBadRequest, body["err"])
	}
}

func TestExchangeForToken_MissingParams(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrMissingParams) {
		t.Fatalf("expected err %q, got %q", ErrMissingParams, body["err"])
	}
}

func TestExchangeForToken_InvalidVerifierFormat(t *testing.T) {
	p, _, _, _ := setupAuthorizer()

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: "abc",
		AuthToken:     "def",
		Verifier:      "short",
	}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrInvalidParam) {
		t.Fatalf("expected err %q, got %q", ErrInvalidParam, body["err"])
	}
}

func TestExchangeForToken_ExchangeTokenMissing(t *testing.T) {
	p, _, _, _ := setupAuthorizer()
	verifier, _ := validVerifierAndChallenge()

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: "missing",
		AuthToken:     "token",
		Verifier:      verifier,
	}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrExchangeFailed) {
		t.Fatalf("expected err %q, got %q", ErrExchangeFailed, body["err"])
	}
}

func TestExchangeForToken_NotClaimed(t *testing.T) {
	p, reg, signer, _ := setupAuthorizer()
	verifier, challenge := validVerifierAndChallenge()

	tokenID := "exchange-123"
	err := reg.SaveExchangeToken(context.Background(), tokenID, lti_domain.ExchangeToken{
		Data:           &lti_domain.SwapToken{},
		ClaimableUntil: time.Now().Add(5 * time.Minute),
		Exchanged:      false,
		Challenge:      "challenge-123",
		AuthToken:      "auth-token",
	}, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: tokenID,
		AuthToken:     "some-auth-token",
		Verifier:      verifier,
	}))
	w := httptest.NewRecorder()

	_ = challenge
	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrExchangeNotClaimed) {
		t.Fatalf("expected err %q, got %q", ErrExchangeNotClaimed, body["err"])
	}

	signer.MustNotHaveSigned(t)
}

func TestExchangeForToken_AuthTokenMismatch(t *testing.T) {
	p, reg, signer, _ := setupAuthorizer()
	verifier, challenge := validVerifierAndChallenge()

	tokenID := "exchange-123"
	err := reg.SaveExchangeToken(context.Background(), tokenID, lti_domain.ExchangeToken{
		Data:           &lti_domain.SwapToken{},
		ClaimableUntil: time.Now().Add(5 * time.Minute),
		Exchanged:      false,
		Challenge:      "challenge-123",
		AuthToken:      "auth-token",
	}, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedAuthToken, err := reg.ClaimExchangeToken(context.Background(), tokenID, challenge)
	if err != nil {
		t.Fatalf("expected no error claiming token, got %v", err)
	}
	if expectedAuthToken == "" {
		t.Fatal("expected auth token from claim, got empty string")
	}

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: tokenID,
		AuthToken:     "wrong-auth-token",
		Verifier:      verifier,
	}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrAuthTokenMismatch) {
		t.Fatalf("expected err %q, got %q", ErrAuthTokenMismatch, body["err"])
	}

	signer.MustNotHaveSigned(t)
}

func TestExchangeForToken_VerifierMismatch(t *testing.T) {
	p, reg, signer, _ := setupAuthorizer()
	verifier, challenge := validVerifierAndChallenge()

	tokenID := "exchange-123"
	err := reg.SaveExchangeToken(context.Background(), tokenID, lti_domain.ExchangeToken{
		Data:           &lti_domain.SwapToken{},
		ClaimableUntil: time.Now().Add(5 * time.Minute),
		Exchanged:      false,
		Challenge:      "challenge-123",
		AuthToken:      "auth-token",
	}, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	authToken, err := reg.ClaimExchangeToken(context.Background(), tokenID, challenge)
	if err != nil {
		t.Fatalf("expected no error claiming token, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: tokenID,
		AuthToken:     authToken,
		Verifier:      verifier + "different",
	}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrVerifierMismatch) {
		t.Fatalf("expected err %q, got %q", ErrVerifierMismatch, body["err"])
	}

	signer.MustNotHaveSigned(t)
}

func TestExchangeForToken_Success(t *testing.T) {
	p, reg, signer, _ := setupAuthorizer()
	verifier, challenge := validVerifierAndChallenge()

	tokenID := "exchange-123"
	err := reg.SaveExchangeToken(context.Background(), tokenID, lti_domain.ExchangeToken{
		Data: &lti_domain.SwapToken{
			To: "/lti/app/finalize",
		},
		ClaimableUntil: time.Now().Add(5 * time.Minute),
		Exchanged:      false,
		Challenge:      "challenge-123",
		AuthToken:      "auth-token",
	}, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	authToken, err := reg.ClaimExchangeToken(context.Background(), tokenID, challenge)
	if err != nil {
		t.Fatalf("expected no error claiming token, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/exchange", mustJSONBody(t, exchRequest{
		ExchangeToken: tokenID,
		AuthToken:     authToken,
		Verifier:      verifier,
	}))
	w := httptest.NewRecorder()

	p.exchangeForToken(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	signer.MustHaveSigned(t)

	body := decodeJSONMap(t, w)
	if body["redirect"] != "/lti/app/finalize" {
		t.Fatalf("expected redirect %q, got %q", "/lti/app/finalize", body["redirect"])
	}

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != lti_domain.ContextKey_Session {
		t.Fatalf("expected cookie name %q, got %q", lti_domain.ContextKey_Session, cookie.Name)
	}
	if cookie.Value != "signed.jwt" {
		t.Fatalf("expected cookie value %q, got %q", "signed.jwt", cookie.Value)
	}
	if cookie.Path != "/lti/app/finalize" {
		t.Fatalf("expected cookie path %q, got %q", "/lti/app/finalize", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected cookie to be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("expected cookie to be Secure")
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("expected SameSite=None, got %v", cookie.SameSite)
	}
}

func TestWriteError_UsesTraceFromContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/exchange", nil).
		WithContext(context.WithValue(context.Background(), "trace_id", "trace-123"))
	w := httptest.NewRecorder()

	writeError(w, req, ErrMissingParams)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Result().StatusCode)
	}

	body := decodeJSONMap(t, w)
	if body["err"] != string(ErrMissingParams) {
		t.Fatalf("expected err %q, got %q", ErrMissingParams, body["err"])
	}
	if body["trace"] != "trace-123" {
		t.Fatalf("expected trace %q, got %q", "trace-123", body["trace"])
	}
}
