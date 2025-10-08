package lti_domain

import "github.com/golang-jwt/jwt/v5"

// LTIJWT represents your internal app-issued JWT after an LTI launch.
// It captures key contextual info for downstream authorization and telemetry.
type LTIJWT struct {
	TenantID    string   `json:"t"`
	Deployment  string   `json:"d"`
	UserID      string   `json:"u"`
	CourseID    string   `json:"c"`
	CourseLabel string   `json:"l"`
	CourseTitle string   `json:"n"`
	Roles       []string `json:"r"`
	jwt.RegisteredClaims
}
