package lti_domain

import "github.com/golang-jwt/jwt/v5"

// LTIJWT represents your internal app-issued JWT after an LTI launch.
// It captures key contextual info for downstream authorization and telemetry.
type LTIJWT struct {
	TenantID   string            `json:"t"`
	Deployment string            `json:"d"`
	ClientID   string            `json:"i"`
	Roles      []Role            `json:"r"`
	UserInfo   LTIJWT_UserInfo   `json:"u"`
	CourseInfo LTIJWT_CourseInfo `json:"c"`
	jwt.RegisteredClaims
}

type LTIJWT_CourseInfo struct {
	CourseID    string `json:"c"`
	CourseLabel string `json:"l"`
	CourseTitle string `json:"n"`
}

type LTIJWT_UserInfo struct {
	UserID     string `json:"u,omitempty"`
	Name       string `json:"n,omitempty"`
	GivenName  string `json:"g,omitempty"`
	FamilyName string `json:"f,omitempty"`
	MiddleName string `json:"m,omitempty"`
	Picture    string `json:"p,omitempty"`
	Email      string `json:"e,omitempty"`
	Locale     string `json:"l,omitempty"`
}
