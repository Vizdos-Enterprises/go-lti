package middleware

import (
	"net/http"
	"slices"

	"github.com/kvizdos/lti-server/lti/lti_domain"
)

func RequireRole(requiredRoles ...lti_domain.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := lti_domain.LTIFromContext(r.Context())
			if !ok {
				http.Error(w, "missing LTI session", http.StatusUnauthorized)
				return
			}

			if len(requiredRoles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			for _, have := range session.Roles {
				if slices.Contains(requiredRoles, have) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "insufficient role", http.StatusForbidden)
		})
	}
}
