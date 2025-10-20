package launcher1dot3

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	"github.com/vizdos-enterprises/go-lti/lti/lti_ports"
)

var _ lti_ports.Launcher = (*LTI13_Launcher)(nil)

type LTI13_Launcher struct {
	registry   lti_ports.Registry
	ephemeral  lti_ports.EphemeralStore
	logger     lti_ports.Logger
	redirector lti_ports.Redirector
	signer     lti_ports.Signer
	keyfunc    lti_ports.KeyfuncProvider

	imposterJWT *lti_domain.LTIJWT

	baseURL  string
	audience []string

	enabledServices []lti_domain.LTIService

	deepLinkingService lti_ports.DeepLinking
}

func (l LTI13_Launcher) GetLTIVersion() string {
	return "1.3"
}

func (l LTI13_Launcher) randomness(length int) (string, error) {
	// Create a byte slice of the desired length.
	b := make([]byte, length)

	// Read random bytes into the slice.
	// crypto/rand.Read returns the number of bytes read and an error if any.
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	// Encode the random bytes into a URL-safe Base64 string.
	// This ensures the string is printable and avoids issues with special characters.
	return base64.URLEncoding.EncodeToString(b), nil
}

func (l LTI13_Launcher) GetAudience() []string {
	return l.audience
}

func (l LTI13_Launcher) HandleOIDC(w http.ResponseWriter, r *http.Request) {
	// Parse form-encoded body
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	deploymentID := r.FormValue("lti_deployment_id")

	deployment, err := l.registry.GetDeployment(r.Context(), clientID, deploymentID)
	if err != nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	iss := r.FormValue("iss")

	if deployment.GetLTIIssuer() != iss {
		l.logger.Error("Invalid issuer", "got", iss, "expected", deployment.GetLTIIssuer())
		http.Error(w, "invalid issuer", http.StatusUnauthorized)
		return
	}

	loginHint := r.FormValue("login_hint")
	targetLink := r.FormValue("target_link_uri")
	messageHint := r.FormValue("lti_message_hint")

	if !strings.HasPrefix(targetLink, fmt.Sprintf("%s/lti/", strings.TrimRight(l.baseURL, "/"))) {
		l.logger.Error("Invalid redirect_uri", "targetLink", targetLink, "expected", fmt.Sprintf("%s/lti/", strings.TrimRight(l.baseURL, "/")))
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	state, err := l.randomness(32)
	if err != nil {
		l.logger.Error("Failed to generate state", "error", err)
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	nonce, err := l.randomness(32)
	if err != nil {
		l.logger.Error("Failed to generate nonce", "error", err)
		http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
		return
	}

	stateData := lti_domain.State{
		Issuer:       iss,
		ClientID:     clientID,
		DeploymentID: deploymentID,
		Nonce:        nonce,
		TenantID:     deployment.GetTenantID(),
		CreatedAt:    time.Now().UTC(),
	}

	err = l.ephemeral.SaveState(r.Context(), state, stateData, 5*time.Minute)
	if err != nil {
		l.logger.Error("Failed to save state, got %s expected %s", err, "nil")
		http.Error(w, "failed to save state", http.StatusInternalServerError)
		return
	}

	v := url.Values{}
	v.Set("response_type", "id_token")
	v.Set("response_mode", "form_post") // required for LTI
	v.Set("client_id", clientID)
	v.Set("scope", "openid")
	v.Set("redirect_uri", targetLink)
	v.Set("login_hint", loginHint)
	if messageHint != "" {
		v.Set("lti_message_hint", messageHint)
	}
	v.Set("state", state)
	v.Set("nonce", nonce)

	redirectURL := fmt.Sprintf("%s?%s", deployment.GetLTIAuthEndpoint(), v.Encode())

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (l LTI13_Launcher) handleImpostering(w http.ResponseWriter, r *http.Request) {
	l.logger.Warn("Impostering Started")
	signed, err := l.signer.Sign(l.imposterJWT, time.Hour)
	if err != nil {
		l.logger.Error("failed to sign internal jwt", "error", err)
		http.Error(w, "internal jwt creation failed", http.StatusInternalServerError)
		return
	}
	l.redirector.RedirectAfterLaunch(w, r, signed)
}

func (l LTI13_Launcher) HandleLaunch(w http.ResponseWriter, r *http.Request) {
	if l.imposterJWT != nil {
		l.handleImpostering(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	rawToken := r.FormValue("id_token")
	stateID := r.FormValue("state")

	if rawToken == "" || stateID == "" {
		http.Error(w, "missing id_token or state", http.StatusBadRequest)
		return
	}

	// Lookup previously saved state
	stateData, err := l.ephemeral.GetState(r.Context(), stateID)
	if err != nil {
		l.logger.Error("Invalid or expired state", "stateID", stateID, "error", err)
		http.Error(w, "invalid or expired state", http.StatusUnauthorized)
		return
	}

	// delete the used state (one-time use)
	_ = l.ephemeral.DeleteState(r.Context(), stateID)

	// Load deployment info (for issuer and JWKS validation)
	dep, err := l.registry.GetDeployment(r.Context(), stateData.ClientID, stateData.DeploymentID)
	if err != nil {
		http.Error(w, "deployment not found", http.StatusUnauthorized)
		return
	}

	// Verify the JWT
	jwksURL := dep.GetLTIJWKSURL()
	k, err := l.keyfunc(r.Context(), []string{jwksURL})
	if err != nil {
		l.logger.Error("Failed to load JWKS", "jwksURL", jwksURL, "error", err)
		http.Error(w, "jwks fetch failed", http.StatusInternalServerError)
		return
	}

	token, err := jwt.Parse(rawToken, k.Keyfunc)
	if err != nil || !token.Valid {
		l.logger.Error("Invalid JWT", "error", err)
		http.Error(w, "invalid id_token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "bad claims", http.StatusInternalServerError)
		return
	}
	if claims["nonce"] != stateData.Nonce {
		l.logger.Error("Invalid nonce used")
		http.Error(w, "invalid nonce", http.StatusUnauthorized)
		return
	}

	messageType, ok := claims["https://purl.imsglobal.org/spec/lti/claim/message_type"].(string)

	requestType := lti_domain.LTIService(messageType)

	if !ok || !slices.Contains(l.enabledServices, requestType) {
		http.Error(w, "invalid message type", http.StatusUnauthorized)
		return
	}

	// Extract relevant fields from the LTI claims
	userID := fmt.Sprintf("%v", claims["sub"])

	ctxClaim, _ := claims["https://purl.imsglobal.org/spec/lti/claim/context"].(map[string]any)
	courseID := fmt.Sprintf("%v", ctxClaim["id"])
	courseLabel := fmt.Sprintf("%v", ctxClaim["label"])
	courseTitle := fmt.Sprintf("%v", ctxClaim["title"])

	rolesRaw, _ := claims["https://purl.imsglobal.org/spec/lti/claim/roles"].([]any)
	roles := make([]lti_domain.Role, 0, len(rolesRaw))
	for _, r := range rolesRaw {
		if str, ok := r.(string); ok {
			roles = append(roles, lti_domain.ParseRoleURI(str))
		}
	}

	customClaims := map[string]any{}
	if custom, ok := claims["https://purl.imsglobal.org/spec/lti/claim/custom"].(map[string]any); ok {
		customClaims = custom
	}

	jwtID, err := l.randomness(16)
	if err != nil {
		l.logger.Error("failed to generate jwt id", "error", err)
		http.Error(w, "jwt id generation failed", http.StatusInternalServerError)
		return
	}

	name, _ := claims["name"].(string)
	given_name, _ := claims["given_name"].(string)
	family_name, _ := claims["family_name"].(string)
	middle_name, _ := claims["middle_name"].(string)
	picture, _ := claims["picture"].(string)
	email, _ := claims["email"].(string)
	locale, _ := claims["locale"].(string)

	fmt.Println(claims)

	// Build your internal JWT payload
	internalClaims := lti_domain.LTIJWT{
		LaunchType: requestType,
		TenantID:   lti_domain.TenantIDString(stateData.TenantID),
		Deployment: dep.GetDeploymentID(),
		ClientID:   dep.GetLTIClientID(),
		Custom:     customClaims,
		CourseInfo: lti_domain.LTIJWT_CourseInfo{
			CourseID:    courseID,
			CourseLabel: courseLabel,
			CourseTitle: courseTitle,
		},
		UserInfo: lti_domain.LTIJWT_UserInfo{
			UserID:     userID,
			Name:       name,
			GivenName:  given_name,
			FamilyName: family_name,
			MiddleName: middle_name,
			Picture:    picture,
			Email:      email,
			Locale:     locale,
		},
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    l.signer.GetIssuer(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Audience:  l.audience,
			ID:        jwtID,
		},
	}

	// Sign it
	signed, err := l.signer.Sign(internalClaims, time.Hour)
	if err != nil {
		l.logger.Error("failed to sign internal jwt", "error", err)
		http.Error(w, "internal jwt creation failed", http.StatusInternalServerError)
		return
	}

	if l.deepLinkingService != nil && l.deepLinkingService.IsDeepLinkLaunch(requestType) {
		l.deepLinkingService.HandleLaunch(w, r, &internalClaims, signed, claims)
		return
	}

	l.redirector.RedirectAfterLaunch(w, r, signed)
}
