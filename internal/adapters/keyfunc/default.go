package keyfunc

import (
	"context"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

func DefaultKeyfuncProviderAdapter() lti_ports.KeyfuncProvider {
	return func(ctx context.Context, urls []string) (lti_ports.Keyfunc, error) {
		kf, err := keyfunc.NewDefaultCtx(ctx, urls)
		if err != nil {
			return nil, err
		}

		return &keyfuncAdapter{kf: kf}, nil
	}
}

type keyfuncAdapter struct {
	kf keyfunc.Keyfunc
}

// Adapt the external library method to your interface.
func (a *keyfuncAdapter) Keyfunc(token *jwt.Token) (any, error) {
	return a.kf.Keyfunc(token)
}
