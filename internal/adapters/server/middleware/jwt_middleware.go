package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/kvizdos/lti-server/lti/lti_domain"
	"github.com/kvizdos/lti-server/lti/lti_ports"
)

func VerifyLTI(verifier lti_ports.Verifier, expectedAudience []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("lti_token")
		if err != nil {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// Parse & verify
		claims := &lti_domain.LTIJWT{}
		token, err := verifier.Verify(cookie.Value, claims)
		if err != nil || !token.Valid {
			fmt.Println("VerifyLTI failed: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Verify audience
		matchFound := false
		aud, err := token.Claims.GetAudience()
		if err != nil {
			http.Error(w, "invalid audience", http.StatusUnauthorized)
			return
		}
		for _, audience := range expectedAudience {
			if slices.Contains(aud, audience) {
				matchFound = true
				break
			}
		}
		if !matchFound {
			http.Error(w, "could not verify audience", http.StatusUnauthorized)
			return
		}

		// Attach to context
		ctx := lti_domain.ContextWithLTI(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
