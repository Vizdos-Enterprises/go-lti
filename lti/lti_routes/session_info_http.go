package lti_routes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
	lti_routes_assets "github.com/vizdos-enterprises/go-lti/lti/lti_routes/assets"
)

type sessionInitializerHTTP struct {
	distinctIdGenerator func(*lti_domain.LTIJWT) string
	tpl                 *template.Template
	minifier            *minify.M
}

func NewSessionInitializerHTTP(distinctIdGenerator func(*lti_domain.LTIJWT) string) *sessionInitializerHTTP {
	tpl := template.Must(template.New("session.js").Parse(string(lti_routes_assets.SessionInitJS)))

	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)

	return &sessionInitializerHTTP{
		distinctIdGenerator: distinctIdGenerator,
		tpl:                 tpl,
		minifier:            m,
	}
}

type sessionInitTemplateData struct {
	UserId         string
	TenantId       string
	RolesJSON      template.HTML
	ContextId      string
	LaunchPlatform string
	Impostering    bool
}

func (s *sessionInitializerHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, ok := lti_domain.LTIFromContext(r.Context())
	if !ok {
		http.Error(w, "Invalid LTI session", http.StatusUnauthorized)
		return
	}

	contextId := "na"
	if session.CourseInfo.CourseID != "" {
		contextId = fmt.Sprintf("course:%s", session.CourseInfo.CourseID)
	}

	rolesJSON, err := json.Marshal(session.Roles)
	if err != nil {
		http.Error(w, "Failed to marshal roles", http.StatusInternalServerError)
		return
	}

	frontend := sessionInitTemplateData{
		UserId:         session.UserInfo.UserID,
		TenantId:       session.TenantID,
		RolesJSON:      template.HTML(rolesJSON),
		ContextId:      contextId,
		LaunchPlatform: session.Platform.Name,
		Impostering:    session.Impostering,
	}

	var rendered bytes.Buffer
	if err := s.tpl.Execute(&rendered, frontend); err != nil {
		http.Error(w, "Failed to render session initializer", http.StatusInternalServerError)
		return
	}

	out, err := s.minifier.Bytes("text/javascript", rendered.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	sum := sha256.Sum256(out)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	if inm := r.Header.Get("If-None-Match"); inm != "" && inm == etag {
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "private, must-revalidate")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "private, must-revalidate")
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(out)

}
