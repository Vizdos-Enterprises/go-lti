package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vizdos-enterprises/go-lti/lti/lti_crypto"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_http"
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

	tenantID := uuid.NewString()
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

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{"4566/tcp"},
		Env: map[string]string{
			"SERVICES": "kms",
			"DEBUG":    "1",
		},
		WaitingFor: wait.ForListeningPort("4566/tcp"),
	}
	localstackC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("failed to start localstack: %v\n", err)
	}
	defer localstackC.Terminate(ctx)

	host, err := localstackC.Host(ctx)
	if err != nil {
		fmt.Printf("failed to get host: %v\n", err)
	}

	mapped, err := localstackC.MappedPort(ctx, "4566/tcp")
	if err != nil {
		fmt.Printf("failed to get mapped port: %v\n", err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, mapped.Port())

	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_REGION", "us-east-1")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Printf("failed to load AWS config: %v\n", err)
	}

	client := kms.NewFromConfig(cfg, func(o *kms.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	// Create KMS key
	createOut, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Description: aws.String("test key"),
		KeyUsage:    types.KeyUsageTypeSignVerify,
		KeySpec:     types.KeySpecEccNistP256,
	})
	if err != nil {
		fmt.Printf("failed to create key: %v\n", err)
	}

	// Init signer
	signer, err := lti_crypto.NewKMS(
		lti_crypto.WithKMS(client, *createOut.KeyMetadata.KeyId, false),
		lti_crypto.WithSigningMethod(jwtkms.SigningMethodECDSA256),
		lti_crypto.WithIssuer("https://dev.kv.codes/"),
	)
	if err != nil {
		fmt.Printf("NewKMS failed: %v\n", err)
	}

	launcher := lti_launcher.NewLTI13Launcher(
		lti_launcher.WithBaseURL(os.Getenv("BASE_URL")),
		lti_launcher.WithRedirectURL("/lti/app"),
		lti_launcher.WithLogger(logger),
		lti_launcher.WithRegistry(registry),
		lti_launcher.WithEphemeralStorage(registry),
		lti_launcher.WithSigner(signer),
		lti_launcher.WithAudience([]string{"made with ❤️ by kenton"}),
	)
	ltiInstance := lti_http.NewServer(
		lti_http.WithLauncher(launcher),
		lti_http.WithVerifier(signer),
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
				Path: "/",
				Role: []lti_domain.Role{lti_domain.MEMBERSHIP_LEARNER},
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
