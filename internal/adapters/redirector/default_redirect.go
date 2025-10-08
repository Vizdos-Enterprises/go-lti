package redirector

import (
	"net/http"

	"github.com/kvizdos/lti-server/lti/lti_ports"
)

var _ lti_ports.Redirector = (*defaultRedirector)(nil)

type defaultRedirector struct {
	redirectURL string
}

func (rw *defaultRedirector) RedirectAfterLaunch(w http.ResponseWriter, r *http.Request, jwt string) {
	cookie := &http.Cookie{
		Name:     "lti_token",
		Value:    jwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, rw.redirectURL, http.StatusFound)
}

func NewDefaultRedirector(baseURL string) lti_ports.Redirector {
	return &defaultRedirector{redirectURL: baseURL}
}
