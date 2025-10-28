package lti_ports

import (
	"net/http"
)

type ImposteringOption func(Impostering)

type Impostering interface {
	HandleImposterLaunch(w http.ResponseWriter, r *http.Request)
}
