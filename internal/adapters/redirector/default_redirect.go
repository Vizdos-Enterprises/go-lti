package redirector

import (
	"net/http"
	"net/url"
	"os"

	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var _ lti_ports.Redirector = (*defaultRedirector)(nil)

type defaultRedirector struct {
	redirectURL string
}

func (rw *defaultRedirector) RedirectAfterLaunch(w http.ResponseWriter, r *http.Request, swapToken string) {
	next, err := url.Parse("/lti/1.3/swap")
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	q := next.Query()
	q.Set("code", swapToken)
	next.RawQuery = q.Encode()

	useSecureCookie := true
	if os.Getenv("INSECURE_COOKIES") == "true" {
		useSecureCookie = false
	}

	cookie := &http.Cookie{
		Name:     lti_domain.ContextKey_CookieConfirmation,
		Value:    swapToken,
		Path:     "/lti/1.3/swap",
		HttpOnly: true,
		Secure:   useSecureCookie,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, next.String(), http.StatusFound)
}

func NewDefaultRedirector(baseURL string) lti_ports.Redirector {
	return &defaultRedirector{redirectURL: baseURL}
}
