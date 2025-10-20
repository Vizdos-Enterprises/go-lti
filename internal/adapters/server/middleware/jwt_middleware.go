package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/deeplinking"
	"github.com/vizdos-enterprises/go-lti/lti/lti_deeplink"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

func parseAndValidate[T jwt.Claims](verifier lti_ports.Verifier, expectedAudience []string, cookieValue string) (*T, error) {
	var claims T // zero value, not a pointer
	token, err := verifier.Verify(cookieValue, any(&claims).(jwt.Claims))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify audience
	matchFound := false
	aud, err := token.Claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("invalid audience")
	}
	for _, audience := range expectedAudience {
		if slices.Contains(aud, audience) {
			matchFound = true
			break
		}
	}
	if len(expectedAudience) == 0 {
		matchFound = true
	}
	if !matchFound {
		return nil, fmt.Errorf("could not verify audience")
	}

	return &claims, nil
}

func VerifyLTI(verifier lti_ports.Verifier, expectedAudience []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(lti_domain.ContextKey_Session)
		if err != nil {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims, err := parseAndValidate[lti_domain.LTIJWT](verifier, expectedAudience, cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		if claims.LaunchType == lti_domain.LTIService_DeepLink {
			deepLinkCookie, err := r.Cookie(deeplinking.ContextKey_DeepLink)
			if err != nil {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			deepLinkContext, err := parseAndValidate[lti_domain.DeepLinkContext](verifier, []string{}, deepLinkCookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if deepLinkContext.AttachedKID != claims.ID {
				http.Error(w, "invalid deep link, mismatch kid", http.StatusUnauthorized)
				return
			}

			ctx = lti_deeplink.ContextWithDeepLink(ctx, deepLinkContext)
			ctx = context.WithValue(ctx, "rawDeepLink", deepLinkCookie)
		}

		// Attach to context
		ctx = lti_domain.ContextWithLTI(ctx, claims)
		ctx = context.WithValue(ctx, "rawJWT", cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
