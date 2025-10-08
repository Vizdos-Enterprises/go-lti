package crypto_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/internal/adapters/crypto"
)

func TestKMSSigner_LocalStackIntegration(t *testing.T) {
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
		t.Fatalf("failed to start localstack: %v", err)
	}
	defer localstackC.Terminate(ctx)

	host, err := localstackC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	mapped, err := localstackC.MappedPort(ctx, "4566/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, mapped.Port())

	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_REGION", "us-east-1")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
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
		t.Fatalf("failed to create key: %v", err)
	}

	// Init signer
	signer, err := crypto.NewKMS(
		crypto.WithKMS(client, *createOut.KeyMetadata.KeyId, false),
		crypto.WithSigningMethod(jwtkms.SigningMethodECDSA256),
		crypto.WithIssuer("kms.test"),
	)
	if err != nil {
		t.Fatalf("NewKMS failed: %v", err)
	}

	// ---- Sign ----
	claims := &jwt.RegisteredClaims{
		Subject: "integration-user",
	}
	tokenStr, err := signer.Sign(claims, time.Minute)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if !strings.Contains(tokenStr, ".") {
		t.Fatalf("expected JWT-like structure, got %s", tokenStr)
	}

	// ---- Verify ----
	parsedClaims := &jwt.RegisteredClaims{}
	token, err := signer.Verify(tokenStr, parsedClaims)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !token.Valid {
		t.Fatal("token was not valid")
	}
	if parsedClaims.Subject != "integration-user" {
		t.Fatalf("expected subject 'integration-user', got %q", parsedClaims.Subject)
	}
	if parsedClaims.Issuer != "kms.test" {
		t.Fatalf("expected issuer 'kms.test', got %q", parsedClaims.Issuer)
	}
}
