package lti_ports

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

type DeepLinkingOption func(DeepLinking)

type DeepLinking interface {
	IsDeepLinkLaunch(lti_domain.LTIService) bool

	HandleLaunch(w http.ResponseWriter, r *http.Request, attachedSession *lti_domain.LTIJWT, jwtString string, claims jwt.MapClaims)
}
