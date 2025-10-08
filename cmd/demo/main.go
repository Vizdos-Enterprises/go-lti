package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kvizdos/lti-server/lti/lti_crypto"
	"github.com/kvizdos/lti-server/lti/lti_domain"
	"github.com/kvizdos/lti-server/lti/lti_http"
	"github.com/kvizdos/lti-server/lti/lti_launcher"
	"github.com/kvizdos/lti-server/lti/lti_logger"
	"github.com/kvizdos/lti-server/lti/lti_registry"
)

func main() {
	_ = godotenv.Load()

	logger := lti_logger.NewSlogLogger()

	registry := lti_registry.NewMemoryRegistry()

	// Demo tenant ID-- tenants are managed outside of this sytem.

	tenantID := lti_domain.TenantID(uuid.New())
	registry.AddDeployment(context.Background(), &lti_domain.Deployment{
		InternalID:    lti_domain.InternalDeploymentID(uuid.New()),
		ForTenantID:   tenantID,
		Issuer:        os.Getenv("LTI_ISSUER"),
		ClientID:      os.Getenv("LTI_CLIENT_ID"),
		JWKSURL:       os.Getenv("LTI_JWKS_URL"),
		AuthEndpoint:  os.Getenv("LTI_AUTH_ENDPOINT"),
		TokenEndpoint: os.Getenv("LTI_TOKEN_ENDPOINT"),
		DeploymentID:  os.Getenv("LTI_DEPLOYMENT_ID"),
	})

	signVerifier := lti_crypto.NewHMAC("kid-demo", "a-string-secret-at-least-256-bits-long", "i-hope-you-have-a-great-day")
	launcher := lti_launcher.NewLTI13Launcher(
		lti_launcher.WithBaseURL(os.Getenv("BASE_URL")),
		lti_launcher.WithRedirectURL("/lti/app"),
		lti_launcher.WithLogger(logger),
		lti_launcher.WithRegistry(registry),
		lti_launcher.WithEphemeralStorage(registry),
		lti_launcher.WithSigner(signVerifier),
		lti_launcher.WithAudience([]string{"made with ❤️ by kenton"}),
	)
	ltiInstance := lti_http.NewServer(
		lti_http.WithLauncher(launcher),
		lti_http.WithVerifier(signVerifier),
	)

	demoRoutes := http.NewServeMux()

	demoRoutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		session, ok := lti_domain.LTIFromContext(r.Context())
		if !ok {
			http.Error(w, "Invalid LTI session", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8" />
				<title>LTI Session Details</title>
				<style>
					body { font-family: sans-serif; margin: 2rem; color: #333; }
					h1 { color: #2b6cb0; }
					table { border-collapse: collapse; width: 100%%; max-width: 600px; margin-top: 1rem; }
					th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
					th { width: 30%%; background: #f7fafc; }
					code { background: #f1f1f1; padding: 2px 4px; border-radius: 4px; }
				</style>
			</head>
			<body>
				<h1>LTI Session Details</h1>
				<table>
					<tr><th>Tenant ID</th><td><code>%s</code></td></tr>
					<tr><th>Deployment</th><td><code>%s</code></td></tr>
					<tr><th>User ID</th><td><code>%s</code></td></tr>
					<tr><th>User Roles</th><td><code>%v</code></td></tr>
					<tr><th>Course ID</th><td><code>%s</code></td></tr>
					<tr><th>Course Label</th><td><code>%s</code></td></tr>
					<tr><th>Course Title</th><td><code>%s</code></td></tr>
					<tr><th>Issued At</th><td><code>%s</code></td></tr>
					<tr><th>Expires At</th><td><code>%s</code></td></tr>
					<tr><th>Issuer</th><td><code>%s</code></td></tr>
					<tr><th>Audience</th><td><code>%v</code></td></tr>
					<tr><th>JWT ID</th><td><code>%s</code></td></tr>
				</table>
			</body>
			</html>
		`,
			session.TenantID,
			session.Deployment,
			session.UserID,
			session.Roles,
			session.CourseID,
			session.CourseLabel,
			session.CourseTitle,
			session.IssuedAt.Time.Format(time.RFC1123),
			session.ExpiresAt.Time.Format(time.RFC1123),
			session.Issuer,
			session.Audience,
			session.ID,
		)
	})

	http.ListenAndServe(":8888", ltiInstance.CreateRoutes(
		lti_http.WithProtectedRoutes(demoRoutes),
	))
}
