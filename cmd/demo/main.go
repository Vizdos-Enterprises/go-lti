package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/vizdos-enterprises/go-lti/lti/lti_crypto"
	"github.com/vizdos-enterprises/go-lti/lti/lti_deeplink"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_http"
	"github.com/vizdos-enterprises/go-lti/lti/lti_impostering"
	"github.com/vizdos-enterprises/go-lti/lti/lti_launcher"
	"github.com/vizdos-enterprises/go-lti/lti/lti_logger"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
	"github.com/vizdos-enterprises/go-lti/lti/lti_registry"
)

func main() {
	_ = godotenv.Load()

	logger := lti_logger.NewSlogLogger()

	registry := lti_registry.NewMemoryRegistry()

	// Demo tenant ID-- tenants are managed outside of this sytem.

	tenantID := "2c24a2a0-5223-47b7-a572-392aac75993a"
	registry.AddDeployment(context.Background(), &lti_domain.BaseLTIDeployment{
		InternalID:    uuid.NewString(),
		ForTenantID:   tenantID,
		Issuer:        os.Getenv("LTI_ISSUER"),
		ClientID:      os.Getenv("LTI_CLIENT_ID"),
		JWKSURL:       os.Getenv("LTI_JWKS_URL"),
		AuthEndpoint:  os.Getenv("LTI_AUTH_ENDPOINT"),
		TokenEndpoint: os.Getenv("LTI_TOKEN_ENDPOINT"),
		DeploymentID:  os.Getenv("LTI_DEPLOYMENT_ID"),
	})
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	signVerifier := lti_crypto.NewRS256("kid-demo", priv, &priv.PublicKey, "https://dev.kv.codes/lti/")

	imposteringSvc := initImpostering(signVerifier)
	deepLinking := lti_deeplink.NewDeepLinkingService(
		lti_deeplink.WithSigner(signVerifier),
		lti_deeplink.WithRedirectURL("/lti/app/deeplink"),
	)

	launcher := lti_launcher.NewLTI13Launcher(
		lti_launcher.WithBaseURL(os.Getenv("BASE_URL")),
		lti_launcher.WithRedirectURL("/lti/app"),
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
		lti_http.WithImpostering(imposteringSvc),
	)

	demoMux := http.NewServeMux()

	demoMux.Handle("/{demo}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := lti_domain.LTIFromContext(r.Context())
		if !ok {
			http.Error(w, "Invalid LTI session", http.StatusUnauthorized)
			return
		}
		fmt.Fprintf(w, "Hello, %s! Your path value is %s", session.UserInfo.Email, r.PathValue("demo"))
	}))

	http.ListenAndServe(":8888", ltiInstance.CreateRoutes(
		lti_http.WithProtectedRoutes(
			lti_ports.ProtectedRoute{
				Path: "/respond",
				Role: []lti_domain.Role{lti_domain.MEMBERSHIP_CONTENT_DEV},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					lti_deeplink.ReplyToDeeplink(w, r, signVerifier, []lti_domain.DeepLinkItem{
						{
							Type:  lti_domain.DeepLinkType_LtiResource,
							Title: "Resource Demo",
							URL:   "https://dev.kv.codes/lti/1.3/launch",
							Custom: map[string]string{
								"app_id": "here",
							},
							Targets: []lti_domain.DeepLinkingTarget{lti_domain.DeepLinkingTarget_Iframe},
							LineItem: &lti_domain.DeepLinkLineItem{
								Label:        "Demo Label",
								ScoreMaximum: 100,
								ResourceID:   "resource-id-demo",
								Tag:          "demo-tag",
							},
						},
					})
				}),
			},
			lti_ports.ProtectedRoute{
				Path: "/deeplink",
				Role: []lti_domain.Role{lti_domain.MEMBERSHIP_CONTENT_DEV},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")

					fmt.Fprintf(w, `
						<!doctype html>
						<html lang="en">
						<head><meta charset="utf-8"><title>Return to LMS</title></head>
						<body style="font-family:sans-serif;display:flex;flex-direction:column;align-items:center;justify-content:center;height:100vh;">
							<h2>Ready to return to your LMS</h2>
							<a href="./respond">Select this item</a>
						</body>
						</html>
					`)
				}),
			},
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
			lti_ports.ProtectedRoute{
				Path:    "/demo/",
				Role:    []lti_domain.Role{lti_domain.MEMBERSHIP_LEARNER},
				Handler: demoMux,
			},
		),
	))
}

func initImpostering(sessionSigner lti_ports.Signer) lti_ports.Impostering {
	verifier := lti_crypto.NewHMAC("incoming", "your-very-secret-key", "demo")
	imp := lti_impostering.NewImpostering(
		lti_impostering.WithSessionSigner(sessionSigner),
		lti_impostering.WithIncomingVerifier(verifier),
		lti_impostering.WithIncomingAudience([]string{"lti-impostering"}),
		lti_impostering.WithSessionAudience([]string{"made with ❤️ by kenton"}),
		lti_impostering.WithLogger(lti_logger.NewSlogLogger()),
	)

	jwtID := uuid.NewString()
	demoJWT := lti_domain.LTIJWT{
		TenantID:   "2c24a2a0-5223-47b7-a572-392aac75993a",
		Deployment: os.Getenv("LTI_DEPLOYMENT_ID"),
		ClientID:   os.Getenv("LTI_CLIENT_ID"),
		Roles: []lti_domain.Role{
			lti_domain.MEMBERSHIP_LEARNER,
		},
		UserInfo: lti_domain.LTIJWT_UserInfo{
			UserID:     "demo-user-id",
			Name:       "Demo User",
			GivenName:  "Demo",
			FamilyName: "User",
			MiddleName: "",
			Picture:    "",
			Email:      "demo@example.com",
			Locale:     "en",
		},
		CourseInfo: lti_domain.LTIJWT_CourseInfo{
			CourseID:    "demo-course-id",
			CourseLabel: "Demo Course",
			CourseTitle: "Demo Course Title",
		},
		LaunchType:             lti_domain.LTIService_ResourceLink,
		Impostering:            true,
		ImposteringSrc:         "demo-user:from-src-app",
		ImposterLaunchRedirect: "/lti/app",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    verifier.GetIssuer(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			Audience:  []string{"lti-impostering"},
			ID:        jwtID,
		},
	}

	signed, err := verifier.Sign(demoJWT, 10*time.Minute)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Demo Impostering Token:\n/lti/imposter?token=%s\n", signed)

	return imp

}
