package launcher1dot3_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

func TestHandleSwap_MissingSwapCode(t *testing.T) {
	l, _, _, signer, _, _ := setupLauncher()

	req := httptest.NewRequest(http.MethodPost, "/swap", nil)
	req.AddCookie(&http.Cookie{
		Name: lti_domain.ContextKey_CookieConfirmation,
	})

	w := httptest.NewRecorder()
	l.HandleCodeSwap(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", resp.StatusCode)
	}

	responseText := w.Body.String()
	if responseText != "missing code\n" {
		t.Fatalf("expected response body %q, got %q", "missing code", responseText)
	}

	signer.MustNotHaveSigned(t)
}

func TestHandleSwap_InvalidSwapCode(t *testing.T) {
	l, _, _, signer, _, _ := setupLauncher()

	req := httptest.NewRequest(http.MethodPost, "/swap", nil)
	q := req.URL.Query()
	q.Add("code", "bad_code")
	req.URL.RawQuery = q.Encode()
	req.AddCookie(&http.Cookie{
		Name: lti_domain.ContextKey_CookieConfirmation,
	})

	w := httptest.NewRecorder()
	l.HandleCodeSwap(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
	}

	responseText := w.Body.String()
	if responseText != "invalid code\n" {
		t.Errorf("expected response body %q, got %q", "invalid code", responseText)
	}

	signer.MustNotHaveSigned(t)
}

func TestHandleSwap_ValidSwapCodeThirdPartyCookie(t *testing.T) {
	l, reg, _, signer, _, _ := setupLauncher()

	tokenID := "demo-exchange-id"
	err := reg.SaveSwapToken(context.Background(), tokenID, lti_domain.SwapToken{
		To:          "/lti/app/finalize",
		RequestorUA: "",
		Claims:      lti_domain.LTIJWT{},
	}, time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/swap", nil)
	q := req.URL.Query()
	q.Add("code", tokenID)
	req.URL.RawQuery = q.Encode()
	req.AddCookie(&http.Cookie{
		Name:  lti_domain.ContextKey_CookieConfirmation,
		Value: tokenID,
	})

	w := httptest.NewRecorder()
	l.HandleCodeSwap(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected %d Redirect, got %d", http.StatusFound, resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if loc != "/lti/app/finalize" {
		t.Errorf("expected Location header %q, got %q", "/lti/app/finalize", loc)
	}

	signer.MustHaveSigned(t)

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Errorf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != lti_domain.ContextKey_Session {
		t.Errorf("expected cookie name %q, got %q", lti_domain.ContextKey_Session, cookie.Name)
	}
	if cookie.Value != "signed.jwt" {
		t.Errorf("expected cookie value %q, got %q", "signed.jwt", cookie.Value)
	}
}

func TestHandleSwap_ValidSwapCodeFallsbackWhenNoCookie(t *testing.T) {
	l, reg, _, signer, _, fallback := setupLauncher()

	tokenID := "demo-exchange-id"
	err := reg.SaveSwapToken(context.Background(), tokenID, lti_domain.SwapToken{
		To:          "/lti/app/finalize",
		RequestorUA: "",
		Claims:      lti_domain.LTIJWT{},
	}, time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/swap", nil)
	q := req.URL.Query()
	q.Add("code", tokenID)
	req.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()
	l.HandleCodeSwap(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("expected %d Redirect, got %d", http.StatusTeapot, resp.StatusCode)
	}

	exchangeID := reg.GetLastSavedExchangeTokenID()

	if exchangeID == "" {
		t.Fatalf("expected exchange token ID to be saved, got empty string")
	}

	signer.MustNotHaveSigned(t)
	fallback.MustHaveBeenCalled(t)
}
