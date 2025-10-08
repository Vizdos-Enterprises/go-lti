package lti_ports

import "net/http"

// Redirector abstracts where to send the user after a successful launch.
type Redirector interface {
	RedirectAfterLaunch(w http.ResponseWriter, r *http.Request, jwt string)
}
