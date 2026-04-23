package helper_routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/helper_routes"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

func TestSessionInfoResponse(t *testing.T) {
	shttp := helper_routes.NewSessionInitializerHTTP(func(l *lti_domain.LTIJWT) string {
		return "demo-id"
	})

	ctx := lti_domain.ContextWithLTI(t.Context(), &lti_domain.LTIJWT{
		TenantID:   "demo-tenant-id",
		Deployment: "demo-deployment-id",
		ClientID:   "demo-client-id",
		Roles:      []lti_domain.Role{lti_domain.INSTITUTION_STUDENT},
		UserInfo: lti_domain.LTIJWT_UserInfo{
			UserID: "demo-user-id",
		},
		CourseInfo: lti_domain.LTIJWT_CourseInfo{
			CourseID: "demo-course-id",
		},
		LaunchType:       lti_domain.LTIService_ResourceLink,
		LinkedResourceID: "demo-resource-id",
		Platform: lti_domain.LTIJWT_ToolPlatform{
			Name: "Demo Platform",
		},
		Impostering: false,
	})

	req := httptest.NewRequest(http.MethodGet, "/session.js", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	shttp.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if got := resp.Header.Get("Content-Type"); got != "application/javascript; charset=utf-8" {
		t.Fatalf("unexpected content-type: %q", got)
	}

	if got := resp.Header.Get("Cache-Control"); got != "private, must-revalidate" {
		t.Fatalf("unexpected cache-control: %q", got)
	}

	if got := resp.Header.Get("ETag"); got == "" {
		t.Fatal("expected ETag header to be set")
	}

	body := rr.Body.String()

	checks := []string{
		"demo-user-id",
		"demo-tenant-id",
		"course:demo-course-id",
		"Demo Platform",
	}

	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got: %s", want, body)
		}
	}

	t.Logf("Output: %s", body)
}
