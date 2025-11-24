package deeplinking

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/redirector"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

const ContextKey_DeepLink = "lti_deep_link"
const ContextKey_DeepLinkSigner string = "lti_deep_link_signer"

var _ lti_ports.DeepLinking = (*DeepLinkingService)(nil)

type DeepLinkingService struct {
	signer     lti_ports.AsymetricSigner
	redirector lti_ports.Redirector
	audience   string
}

func (DeepLinkingService) IsDeepLinkLaunch(status lti_domain.LTIService) bool {
	return status == lti_domain.LTIService_DeepLink
}

func (l DeepLinkingService) randomness(length int) (string, error) {
	// Create a byte slice of the desired length.
	b := make([]byte, length)

	// Read random bytes into the slice.
	// crypto/rand.Read returns the number of bytes read and an error if any.
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	// Encode the random bytes into a URL-safe Base64 string.
	// This ensures the string is printable and avoids issues with special characters.
	return base64.URLEncoding.EncodeToString(b), nil
}

func (d DeepLinkingService) HandleLaunch(w http.ResponseWriter, r *http.Request, attachedSession *lti_domain.LTIJWT, jwtString string, claims jwt.MapClaims) {
	// Extract deep linking settings claim
	rawSettings, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"].(map[string]any)
	if !ok {
		http.Error(w, "missing deep_linking_settings", http.StatusBadRequest)
		return
	}

	var (
		returnURL   string
		data        string
		acceptTypes []lti_domain.DeepLinkType
		targets     []lti_domain.DeepLinkingTarget
		autoCreate  bool
		mediaTypes  string
	)

	if v, ok := rawSettings["deep_link_return_url"].(string); ok {
		returnURL = v
	}
	if v, ok := rawSettings["data"].(string); ok {
		data = v
	}
	if v, ok := rawSettings["accept_types"].([]any); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				acceptTypes = append(acceptTypes, lti_domain.DeepLinkType(s))
			}
		}
	}
	if v, ok := rawSettings["accept_presentation_document_targets"].([]any); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				targets = append(targets, lti_domain.DeepLinkingTarget(s))
			}
		}
	}
	if v, ok := rawSettings["auto_create"].(bool); ok {
		autoCreate = v
	}
	if v, ok := rawSettings["accept_media_types"].(string); ok {
		mediaTypes = v
	}

	jwtID, err := d.randomness(16)
	if err != nil {
		http.Error(w, "jwt id generation failed", http.StatusInternalServerError)
		return
	}

	deepLinkContext := &lti_domain.DeepLinkContext{
		ReturnURL:        returnURL,
		ReturnAud:        claims["iss"].(string),
		Nonce:            claims["nonce"].(string),
		Data:             data,
		AcceptTypes:      acceptTypes,
		Targets:          targets,
		AutoCreate:       autoCreate,
		AcceptMediaTypes: mediaTypes,
		AttachedKID:      attachedSession.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    d.signer.GetIssuer(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-10 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			ID:        jwtID,
		},
	}

	deepLinkContextJWT, err := d.signer.Sign(deepLinkContext, 10*time.Minute)

	if err != nil {
		http.Error(w, "failed to sign deep link context", http.StatusInternalServerError)
		return
	}

	deepLinkContextCookie := &http.Cookie{
		Name:     ContextKey_DeepLink,
		Value:    deepLinkContextJWT,
		Path:     "/lti/app/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, deepLinkContextCookie)

	d.redirector.RedirectAfterLaunch(w, r, jwtString)
}

func NewDeepLinkingService(opts ...lti_ports.DeepLinkingOption) lti_ports.DeepLinking {
	svc := &DeepLinkingService{}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func WithSigner(signer lti_ports.AsymetricSigner) lti_ports.DeepLinkingOption {
	return func(s lti_ports.DeepLinking) {
		svc := s.(*DeepLinkingService)
		svc.signer = signer
	}
}

func WithRedirectURL(url string) lti_ports.DeepLinkingOption {
	return func(s lti_ports.DeepLinking) {
		svc := s.(*DeepLinkingService)
		svc.redirector = redirector.NewDefaultRedirector(url)
	}
}
