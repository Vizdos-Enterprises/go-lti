# go-lti Framework

A lightweight, hexagonal Go framework for building LTI 1.3-compliant servers that integrate with any LMS (e.g., Agilix Buzz, Canvas, Schoology).

Built for real-world deployment, it includes full support for Safari and cookieless contexts, handling third-party cookie restrictions seamlessly through a secure fallback flow so launches remain reliable across all browsers without extra work.*

Designed to be modular, multi-tenant ready, and easy to embed, it lets you go from concept to a production-ready LTI integration in minutes.

**AGS Coming Soon**

*Currently uses a custom workaround; planned support for PostMessage methods later.

## Architecture

- Hexagonal design — clean separation between domain, ports, and adapters.
- Multi-tenant ready — registry interface allows database or in-memory backends.
- Pluggable — swap in any port easily, or make your own.
- Embeddable — start an LTI-compliant server in any Go service via package import.

## Directory Overview
```
lti/
  lti_ports      // hexagonal ports for adapters defined
  lti_crypto     // signing & verification
  lti_domain     // core types and session state
  lti_http       // LTI server and middleware
  lti_launcher   // OIDC + LTI 1.3 launch handler
  lti_logger     // pluggable logger
  lti_registry   // in-memory registry
```

## Demo

First, copy the `.env.example` file to `.env` and update the values as needed.

```
go run ./cmd/demo
```

This will start a demo server. In your LMS, add a new Tool:

```
OpenID Connect Login URL: https://<proxied domain>/lti/1.3/oidc
Tool Key Set URL: https://<proxied domain>/lti/keys.json <coming soon!>
LTI Tool Redirect: https://<proxied domain>/lti/1.3/launch
```

Give it a spin! You should see a page showing the *application* JWT decoded (this is different from the LTI JWT).

## Example

```go
registry := lti_registry.NewMemoryRegistry()
signer := lti_crypto.NewHMAC("super-secret", "issuer", time.Hour) // time.Hour is the duration of the JWT

launcher := lti_launcher.NewLTI13Launcher(
    lti_launcher.WithBaseURL("https://your-domain.com"),
    lti_launcher.WithRedirectURL("/lti/app"),
    lti_launcher.WithRegistry(registry),
    lti_launcher.WithSigner(signer),
)

server := lti_http.NewServer(
    lti_http.WithLauncher(launcher),
    lti_http.WithVerifier(signer),
)

demoRoutes := http.NewServeMux()

demoRoutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	session, _ := lti_domain.LTIFromContext(r.Context())
	fmt.Fprintf(w, "Hello, %s!", session.UserID)
})

http.ListenAndServe(":8888", server.CreateRoutes(
    lti_http.WithProtectedRoutes(demoRoutes), // This will be established at /lti/app/<routes>. It will pass through JWT verification prior to routing.
))
```

## Roadmap

- [x] JWKS Endpoint
- [x] Role-based Authorization
- [x] AWS KMS JWT Provider
- [x] Deep Linking
- [ ] AGS (Assignment & Grade Service)
- [ ] Pluggable Storage Providers - base adapters for: MongoDB (persistent registry), PostgreSQL (persistent registry), Redis (ephemeral state)
- [ ] NRPS (Names and Role Provisioning Services)
- [ ] Telemtry & Metrics - Structured tracing (OpenTelemetry) and metrics for launch latency, success rate, and platform distribution.
