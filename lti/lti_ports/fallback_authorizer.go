package lti_ports

import (
	"net/http"
)

type FallbackAuthorizer interface {
	HandleFallback(w http.ResponseWriter, r *http.Request, exchangeToken string)
	Route() *http.ServeMux
}
