package lti_ports

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Launcher interface {
	GetLTIVersion() string // e.g. "1.3"
	GetAudience() []string
	HandleOIDC(w http.ResponseWriter, r *http.Request)
	HandleLaunch(w http.ResponseWriter, r *http.Request)
}

type Keyfunc interface {
	Keyfunc(token *jwt.Token) (any, error)
}

type KeyfuncProvider func(ctx context.Context, urls []string) (Keyfunc, error)
