package demo_lms

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type DemoPlatform struct {
	Issuer       string
	ClientID     string
	DeploymentID string

	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	KeyID      string
}

func NewDemoPlatform(baseURL string) *DemoPlatform {
	pub, priv, _ := ed25519.GenerateKey(nil)
	return &DemoPlatform{
		Issuer:       baseURL,
		ClientID:     "demo-client",
		DeploymentID: "demo-deployment",
		PrivateKey:   priv,
		PublicKey:    pub,
		KeyID:        "demo-key-1",
	}
}

func (d *DemoPlatform) JWKSURL() string {
	return d.Issuer + "/jwks.json"
}

func (d *DemoPlatform) Routes(toolBaseURL string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/jwks.json", d.handleJWKS)

	mux.HandleFunc("/render", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		html := `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Render</title>
  <style>
    html, body {
      margin: 0;
      padding: 0;
      height: 100%;
    }
    iframe {
      border: none;
      width: 100%;
      height: 100%;
      display: block;
    }
  </style>
</head>
<body>
  <iframe src="/start"></iframe>
</body>
</html>`

		w.Write([]byte(html))
	})

	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		target := fmt.Sprintf(
			"%s/lti/1.3/oidc?iss=%s&client_id=%s&login_hint=test-login&lti_message_hint=test-msg&target_link_uri=%s/lti/1.3/launch&lti_deployment_id=%s",
			toolBaseURL,
			d.Issuer,
			d.ClientID,
			toolBaseURL,
			d.DeploymentID,
		)
		http.Redirect(w, r, target, http.StatusFound)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		nonce := r.URL.Query().Get("nonce")

		idToken, err := d.signLaunchToken(toolBaseURL, nonce)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("post").Parse(`
<!doctype html>
<html>
<body onload="document.forms[0].submit()">
<form method="POST" action="{{.Action}}">
  <input type="hidden" name="id_token" value="{{.IDToken}}">
  <input type="hidden" name="state" value="{{.State}}">
</form>
</body>
</html>`))

		_ = tmpl.Execute(w, map[string]string{
			"Action":  toolBaseURL + "/lti/1.3/launch",
			"IDToken": idToken,
			"State":   state,
		})
	})

	return mux
}

func (d *DemoPlatform) signLaunchToken(toolBaseURL, nonce string) (string, error) {
	claims := jwt.MapClaims{
		"iss":   d.Issuer,
		"aud":   d.ClientID,
		"sub":   "student-123",
		"exp":   time.Now().Add(5 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": nonce,

		"https://purl.imsglobal.org/spec/lti/claim/message_type":  "LtiResourceLinkRequest",
		"https://purl.imsglobal.org/spec/lti/claim/version":       "1.3.0",
		"https://purl.imsglobal.org/spec/lti/claim/deployment_id": d.DeploymentID,
		"https://purl.imsglobal.org/spec/lti/claim/roles": []string{
			"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner",
		},
		"https://purl.imsglobal.org/spec/lti/claim/context": map[string]any{
			"id":    "course-1",
			"label": "BIO101",
			"title": "Biology 101",
		},
		"https://purl.imsglobal.org/spec/lti/claim/resource_link": map[string]any{
			"id": "resource-1",
		},
		"https://purl.imsglobal.org/spec/lti/claim/tool_platform": map[string]any{
			"guid":                "demo-platform",
			"name":                "Demo LMS",
			"product_family_code": "demo",
			"url":                 d.Issuer,
			"version":             "1.0",
		},
		"name":        "Test Student",
		"given_name":  "Test",
		"family_name": "Student",
		"email":       "student@example.com",
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tok.Header["kid"] = d.KeyID
	return tok.SignedString(d.PrivateKey)
}

func (d *DemoPlatform) handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pub, ok := any(d.PublicKey).(ed25519.PublicKey)
	if !ok {
		http.Error(w, "public key is not ed25519", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "OKP",
				"crv": "Ed25519",
				"alg": "EdDSA",
				"use": "sig",
				"kid": d.KeyID,
				"x":   base64.RawURLEncoding.EncodeToString(pub),
			},
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
