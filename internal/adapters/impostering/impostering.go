package impostering

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var _ lti_ports.Impostering = (*ImposteringService)(nil)

type ImposteringService struct {
	incomingVerifier lti_ports.Verifier
	sessionSigner    lti_ports.Signer
	audience         []string
	sessionAud       []string

	logger lti_ports.Logger
}

func (s *ImposteringService) HandleImposterLaunch(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")

	if tokenStr == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	var jwt lti_domain.LTIJWT
	token, err := s.incomingVerifier.Verify(tokenStr, &jwt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if jwt.ImposterLaunchRedirect == "" {
		http.Error(w, "imposter launch redirect is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(jwt.ImposterLaunchRedirect, "/lti/app") {
		http.Error(w, "invalid imposter launch redirect", http.StatusBadRequest)
		return
	}

	aud, err := token.Claims.GetAudience()
	if err != nil {
		http.Error(w, "invalid audience", http.StatusBadRequest)
		return
	}
	matchFound := false
	for _, audience := range s.audience {
		if slices.Contains(aud, audience) {
			matchFound = true
			break
		}
	}

	if !matchFound {
		http.Error(w, "unexpected audience", http.StatusBadRequest)
		return
	}

	if !jwt.Impostering {
		http.Error(w, "not an imposter token", http.StatusUnauthorized)
		return
	}

	if jwt.ImposteringSrc == "" {
		http.Error(w, "imposter source is required", http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	fmt.Println(jwt.ImposterLaunchRedirect)

	redirect := jwt.ImposterLaunchRedirect
	jwt.Audience = s.sessionAud
	jwt.ID = uuid.New().String()
	jwt.ImposterLaunchRedirect = ""
	signed, err := s.sessionSigner.Sign(jwt, time.Hour)
	if err != nil {
		s.logger.Error("failed to sign internal jwt for impostering session", "error", err)
		http.Error(w, "internal jwt creation failed", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     lti_domain.ContextKey_Session,
		Value:    signed,
		Path:     "/lti/app/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	s.logger.Info("impostering session started", "src", jwt.ImposteringSrc, "for_user", jwt.UserInfo.UserID, "impostering_id", jwt.ID, "redirect", redirect)

	http.SetCookie(w, cookie)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *ImposteringService) Authorize(token string) error {
	return nil
}
