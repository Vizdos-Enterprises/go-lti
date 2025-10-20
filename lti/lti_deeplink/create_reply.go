package lti_deeplink

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	deeplinking_html "github.com/vizdos-enterprises/go-lti/internal/adapters/deeplinking/html"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func CreateReplyJWT(signer lti_ports.AsymetricSigner, ctx *lti_domain.DeepLinkContext, session *lti_domain.LTIJWT, items []lti_domain.DeepLinkItem) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":   signer.GetIssuer(), // your tool’s registered issuer
		"aud":   ctx.ReturnAud,      // Buzz client_id from DeepLinkRequest
		"iat":   now.Unix(),
		"exp":   now.Add(1 * time.Minute).Unix(),
		"jti":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"nonce": ctx.Nonce,

		// required LTI claims
		"https://purl.imsglobal.org/spec/lti/claim/deployment_id": session.Deployment,
		"https://purl.imsglobal.org/spec/lti/claim/message_type":  "LtiDeepLinkingResponse",
		"https://purl.imsglobal.org/spec/lti/claim/version":       "1.3.0",

		// deep-link-specific claims
		"https://purl.imsglobal.org/spec/lti-dl/claim/content_items": items,
		"https://purl.imsglobal.org/spec/lti-dl/claim/data":          ctx.Data,
	}

	token, err := signer.Sign(claims, 5*time.Minute)
	return token, err
}

func ReplyToDeeplink(w http.ResponseWriter, r *http.Request, signer lti_ports.AsymetricSigner, items []lti_domain.DeepLinkItem) error {
	session, _ := lti_domain.LTIFromContext(r.Context())
	deepLinkContext, _ := DeepLinkFromContext(r.Context())
	responseJWT, err := CreateReplyJWT(signer, deepLinkContext, session, items)
	if err != nil {
		http.Error(w, "failed to generate JWT", http.StatusInternalServerError)
		return err
	}

	data := struct {
		ReturnURL string
		JWT       string
	}{
		ReturnURL: deepLinkContext.ReturnURL,
		JWT:       responseJWT,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := deeplinking_html.ReturnToLMSHTML.Execute(w, data); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
		return err
	}

	return nil
}
