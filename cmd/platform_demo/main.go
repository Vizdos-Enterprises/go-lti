package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/vizdos-enterprises/go-lti/internal/demo_lms"
	"github.com/vizdos-enterprises/go-lti/lti/lti_crypto"
	"github.com/vizdos-enterprises/go-lti/lti/lti_deeplink"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_http"
	"github.com/vizdos-enterprises/go-lti/lti/lti_launcher"
	"github.com/vizdos-enterprises/go-lti/lti/lti_logger"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
	"github.com/vizdos-enterprises/go-lti/lti/lti_registry"
)

const (
	lmsHost  = "platform.127-0-0-1.sslip.io:9999"
	toolHost = "tool.127.0.0.1.nip.io:9898"
)

func main() {
	// os.Setenv("INSECURE_COOKIES", "true")
	_ = godotenv.Load()

	logger := lti_logger.NewSlogLogger()
	registry := lti_registry.NewMemoryRegistry()

	platform := demo_lms.NewDemoPlatform("https://" + lmsHost)

	tenantID := "2c24a2a0-5223-47b7-a572-392aac75993a"

	registry.AddDeployment(context.Background(), &lti_domain.BaseLTIDeployment{
		InternalID:    uuid.NewString(),
		ForTenantID:   tenantID,
		Issuer:        "https://" + lmsHost,
		ClientID:      "demo-client",
		JWKSURL:       "https://" + lmsHost + "/jwks.json",
		AuthEndpoint:  "https://" + lmsHost + "/auth",
		TokenEndpoint: "https://" + lmsHost + "/token",
		DeploymentID:  "demo-deployment",
	})

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	signVerifier := lti_crypto.NewRS256(
		"kid-demo",
		priv,
		&priv.PublicKey,
		"https://"+toolHost+"/lti/",
	)

	deepLinking := lti_deeplink.NewDeepLinkingService(
		lti_deeplink.WithSigner(signVerifier),
		lti_deeplink.WithRedirectURL("/lti/app/deeplink"),
	)

	launcher := lti_launcher.NewLTI13Launcher(
		lti_launcher.WithBaseURL("https://"+toolHost),
		lti_launcher.WithRedirectURL("/lti/app/"),
		lti_launcher.WithLogger(logger),
		lti_launcher.WithRegistry(registry),
		lti_launcher.WithEphemeralStorage(registry),
		lti_launcher.WithSigner(signVerifier),
		lti_launcher.WithAudience([]string{"made with ❤️ by kenton"}),
		lti_launcher.WithDeepLinking(deepLinking),
	)

	ltiInstance := lti_http.NewServer(
		lti_http.WithLauncher(launcher),
		lti_http.WithVerifier(signVerifier),
	)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http.DefaultClient = &http.Client{
		Transport: tr,
	}
	go func() {
		log.Fatal(http.ListenAndServeTLS(
			":9999",
			"certs/platform.pem",
			"certs/platform-key.pem",
			platform.Routes("https://"+toolHost),
		))
	}()

	if err := http.ListenAndServeTLS(
		":9898",
		"certs/tool.pem",
		"certs/tool-key.pem",
		ltiInstance.CreateRoutes(
			lti_http.WithProtectedRoutes(
				lti_ports.ProtectedRoute{
					Path:             "/",
					Role:             []lti_domain.Role{lti_domain.MEMBERSHIP_LEARNER},
					AllowImpostering: true,
					Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						session, ok := lti_domain.LTIFromContext(r.Context())
						if !ok {
							http.Error(w, "Invalid LTI session", http.StatusUnauthorized)
							return
						}

						rawToken := r.Context().Value("rawJWT")

						w.Header().Set("Content-Type", "text/html; charset=utf-8")

						fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<title>LTI JWT Details</title>
<style>
	body { font-family: sans-serif; margin: 2rem; color: #333; }
	h1 { color: #2b6cb0; }
	table { border-collapse: collapse; width: 100%%; max-width: 800px; }
	th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; vertical-align: top; }
	th { width: 30%%; background: #f7fafc; }
	code { background: #f1f1f1; padding: 2px 4px; border-radius: 4px; }
</style>
</head>
<body>
<h1>LTI JWT – Full Field Breakdown</h1>
<table>
	<tr><th>Tenant ID</th><td><code>%s</code></td></tr>
	<tr><th>Deployment ID</th><td><code>%s</code></td></tr>
	<tr><th>User ID</th><td><code>%s</code></td></tr>
	<tr><th>User Full Name</th><td><code>%s</code></td></tr>
	<tr><th>Given Name</th><td><code>%s</code></td></tr>
	<tr><th>Family Name</th><td><code>%s</code></td></tr>
	<tr><th>Middle Name</th><td><code>%s</code></td></tr>
	<tr><th>Profile Picture URL</th><td><code>%s</code></td></tr>
	<tr><th>Email</th><td><code>%s</code></td></tr>
	<tr><th>Locale</th><td><code>%s</code></td></tr>

	<tr><th>Roles</th><td><code>%v</code></td></tr>

	<tr><th>Course ID</th><td><code>%s</code></td></tr>
	<tr><th>Course Label</th><td><code>%s</code></td></tr>
	<tr><th>Course Title</th><td><code>%s</code></td></tr>

	<tr><th>Issuer</th><td><code>%s</code></td></tr>
	<tr><th>Audience</th><td><code>%v</code></td></tr>
	<tr><th>JWT ID</th><td><code>%s</code></td></tr>
	<tr><th>Issued At</th><td><code>%s</code></td></tr>
	<tr><th>Expires At</th><td><code>%s</code></td></tr>
	<tr><th>Not Before</th><td><code>%s</code></td></tr>
	<tr><th>Custom Claims</th><td><code>%+v</code></td></tr>
	<tr><th>Impostering</th><td><code>%s</code></td></tr>
	<tr><th>Impostering Source</th><td><code>%s</code></td></tr>
	<tr><th>Raw Token</th><td><code>%s</code></td></tr>
</table>
</body>
</html>
`,
							session.TenantID,
							session.Deployment,
							session.UserInfo.UserID,
							session.UserInfo.Name,
							session.UserInfo.GivenName,
							session.UserInfo.FamilyName,
							session.UserInfo.MiddleName,
							session.UserInfo.Picture,
							session.UserInfo.Email,
							session.UserInfo.Locale,
							session.Roles,
							session.CourseInfo.CourseID,
							session.CourseInfo.CourseLabel,
							session.CourseInfo.CourseTitle,
							session.Issuer,
							session.Audience,
							session.ID,
							session.IssuedAt,
							session.ExpiresAt,
							session.NotBefore,
							session.Custom,
							strconv.FormatBool(session.Impostering),
							session.ImposteringSrc,
							rawToken,
						)
					}),
				},
			),
		)); err != nil {
		panic(err)
	}
}
