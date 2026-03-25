package fallback_authorizer

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	pages "github.com/vizdos-enterprises/go-lti/internal/adapters/fallback_authorizer/frontend"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/observability"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

type pkceAuthorizer struct {
	ephemeral lti_ports.EphemeralStore
	signer    lti_ports.Signer
	logger    lti_ports.Logger
	telemetry lti_ports.TelemetryPort
}

func New(store lti_ports.EphemeralStore, signer lti_ports.Signer, logger lti_ports.Logger, telemetry lti_ports.TelemetryPort) *pkceAuthorizer {
	return &pkceAuthorizer{ephemeral: store, signer: signer, logger: logger, telemetry: telemetry}
}

func (p *pkceAuthorizer) HandleFallback(w http.ResponseWriter, r *http.Request, exchangeToken string) {
	joined := fmt.Sprintf("/lti/auth/verify?exchange=%s", exchangeToken)
	http.Redirect(w, r, joined, http.StatusFound)
}

func (p *pkceAuthorizer) Route() *http.ServeMux {
	pagesMux := http.NewServeMux()
	pagesMux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		w.Write(pages.Styles)
	})

	pagesMux.HandleFunc("/init", p.initExchangeCode)
	pagesMux.HandleFunc("/exchange", p.exchangeForToken)

	pagesMux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(pages.InterstitialHTML)
	})

	pagesMux.HandleFunc("/continue", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(pages.ExchangeHTML)
	})

	safelyIgnoredErrors := []string{
		"missing token",
		"invalid token",
		"role",
	}

	pagesMux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		err := r.URL.Query().Get("err")
		if !slices.Contains(safelyIgnoredErrors, err) {
			p.logger.Error("showing error page", "err", err)
		} else {
			p.logger.Warn("showing error page with a safely ignored error", "err", err)
		}
		w.Header().Add("Content-Type", "text/html")
		w.Write(pages.ErrorHTML)
	})

	return pagesMux
}

type exchRequest struct {
	ExchangeToken string `json:"exchange"`
	Verifier      string `json:"verifier"`
	AuthToken     string `json:"token"`
}

type requestErr string

const (
	ErrMethodNotAllowed   requestErr = "H96-IF4"
	ErrBadRequest         requestErr = "NOL-K1S"
	ErrFailedToExchange   requestErr = "STR-UH3"
	ErrMissingParams      requestErr = "VAA-JQT"
	ErrInvalidParam       requestErr = "1W1-V2S"
	ErrExchangeFailed     requestErr = "U6X-ILR"
	ErrExchangeNotClaimed requestErr = "U6Z-ILR"
	ErrAuthTokenMismatch  requestErr = "C9J-1QU"
	ErrVerifierMismatch   requestErr = "0MB-WIU"
	ErrFailedToSign       requestErr = "OYF-JQN"
)

func writeError(w http.ResponseWriter, r *http.Request, errorCode requestErr) {
	trace, ok := r.Context().Value("trace_id").(string)
	if !ok {
		trace = rand.Text()
	}
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"err":   string(errorCode),
		"trace": trace,
	})
}

func (p *pkceAuthorizer) logFromContext(r *http.Request) []any {
	out := []any{"method", r.Method}

	if traceID, ok := r.Context().Value("trace_id").(string); ok {
		out = append(out, "trace", traceID)
	}

	if requestID, ok := r.Context().Value("request_id").(string); ok {
		out = append(out, "request", requestID)
	}

	return out
}

func (p *pkceAuthorizer) withContext(r *http.Request, kv ...any) []any {
	return append(kv, p.logFromContext(r)...)
}

func (p *pkceAuthorizer) exchangeForToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		p.logger.Warn("method not allowed", p.withContext(r)...)
		writeError(w, r, ErrMethodNotAllowed)
		return
	}

	var req exchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		observability.CaptureRequestError(r, err, "decode exchange request", "handler", "exchangeForToken")
		p.logger.Error("bad request", p.withContext(r, "error", err.Error())...)
		writeError(w, r, ErrBadRequest)
		return
	}

	if req.AuthToken == "" || req.ExchangeToken == "" || req.Verifier == "" {
		p.logger.Warn("missing params", p.withContext(r)...)
		writeError(w, r, ErrMissingParams)
		return
	}

	var pkceVerifierRE = regexp.MustCompile(`^[A-Za-z0-9\-._~]{43,128}$`)
	if !pkceVerifierRE.MatchString(req.Verifier) {
		p.logger.Warn("invalid params", p.withContext(r)...)
		writeError(w, r, ErrInvalidParam)
		return
	}

	exchangeInfo, err := p.ephemeral.GetAndDeleteExchangeToken(r.Context(), req.ExchangeToken)
	if err != nil {
		observability.CaptureRequestError(r, err, "exchange token lookup failed", "exchange_token", req.ExchangeToken)
		p.logger.Error("failed to GetAndDeleteExchangeToken", p.withContext(r, "error", err, "code", ErrExchangeFailed)...)
		writeError(w, r, ErrExchangeFailed)
		return
	}

	if !exchangeInfo.Exchanged {
		p.logger.Error("token was attempted to be exchanged but was not claimed", p.withContext(r, "code", ErrExchangeNotClaimed)...)
		writeError(w, r, ErrExchangeNotClaimed)
		return
	}

	if subtle.ConstantTimeCompare([]byte(exchangeInfo.AuthToken), []byte(req.AuthToken)) != 1 {
		p.logger.Error("token mismatch occurred", p.withContext(r, "code", ErrAuthTokenMismatch)...)
		writeError(w, r, ErrAuthTokenMismatch)
		return
	}

	sum := sha256.Sum256([]byte(req.Verifier))
	verifierChallenge := base64.RawURLEncoding.EncodeToString(sum[:])

	if subtle.ConstantTimeCompare(
		[]byte(exchangeInfo.Challenge),
		[]byte(verifierChallenge),
	) != 1 {
		p.logger.Error("verifier mismatch occurred", p.withContext(r, "code", ErrVerifierMismatch)...)
		writeError(w, r, ErrVerifierMismatch)
		return
	}

	signed, err := p.signer.Sign(exchangeInfo.Data.Claims, time.Hour)
	if err != nil {
		observability.CaptureRequestError(r, err, "failed to sign internal jwt")
		p.logger.Error("failed to sign internal jwt", p.withContext(r, "error", err, "code", ErrFailedToSign)...)
		writeError(w, r, ErrFailedToSign)
		return
	}

	useSecureCookie := true
	if os.Getenv("INSECURE_COOKIES") == "true" {
		useSecureCookie = false
	}
	p.telemetry.EmitLaunch(lti_domain.LaunchEvent{
		At:          time.Now().UTC(),
		Method:      lti_domain.LaunchMethodPKCE,
		Success:     true,
		Platform:    exchangeInfo.Data.Claims.Platform.ProductFamilyCode,
		UserAgent:   exchangeInfo.Data.RequestorUA,
		Duration:    time.Since(exchangeInfo.Data.StartAt),
		UserID:      exchangeInfo.Data.Claims.UserInfo.UserID,
		Impostering: exchangeInfo.Data.Claims.Impostering,
	})
	cookie := &http.Cookie{
		Name:     lti_domain.ContextKey_Session,
		Value:    signed,
		Path:     exchangeInfo.Data.To,
		HttpOnly: true,
		Secure:   useSecureCookie,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"redirect": exchangeInfo.Data.To,
	})
}

type initRequest struct {
	Challenge     string `json:"challenge"`
	ExchangeToken string `json:"exchange_token"`
}

func (p *pkceAuthorizer) initExchangeCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		p.logger.Warn("method not allowed", p.withContext(r, "code", ErrMethodNotAllowed)...)
		writeError(w, r, ErrMethodNotAllowed)
		return
	}

	var req initRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.logger.Error("bad request", p.withContext(r, "error", err.Error(), "code", ErrBadRequest)...)
		writeError(w, r, ErrBadRequest)
		return
	}

	exchangeToken := req.ExchangeToken
	challenge := req.Challenge

	if exchangeToken == "" || challenge == "" {
		p.logger.Warn("missing params", p.withContext(r, "code", ErrMissingParams)...)
		writeError(w, r, ErrMissingParams)
		return
	}

	authToken, err := p.ephemeral.ClaimExchangeToken(r.Context(), exchangeToken, challenge)
	if err != nil {
		if errors.Is(err, lti_domain.ErrExchangeTokenNotFound) {
			p.logger.Warn("exchange token not found", p.withContext(r, "code", ErrFailedToExchange)...)
		} else {
			p.logger.Error("failed to claim exchange token", p.withContext(r, "error", err.Error(), "code", ErrFailedToExchange)...)
		}
		writeError(w, r, ErrFailedToExchange)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": authToken,
	})
}
