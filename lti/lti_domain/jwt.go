package lti_domain

import (
	"github.com/golang-jwt/jwt/v5"
)

// LTIJWT represents your internal app-issued JWT after an LTI launch.
// It captures key contextual info for downstream authorization and telemetry.
type LTIJWT struct {
	TenantID               string              `json:"t"`
	Deployment             string              `json:"d"`
	ClientID               string              `json:"i"`
	Roles                  []Role              `json:"r"`
	UserInfo               LTIJWT_UserInfo     `json:"u"`
	CourseInfo             LTIJWT_CourseInfo   `json:"c"`
	LaunchType             LTIService          `json:"s"`
	LinkedResourceID       string              `json:"lr"`
	Platform               LTIJWT_ToolPlatform `json:"p"`
	Custom                 map[string]any      `json:"cu"`
	Impostering            bool                `json:"im"`
	ImposteringSrc         string              `json:"ims,omitempty"`
	ImposterLaunchRedirect string              `json:"ilr,omitempty"`
	jwt.RegisteredClaims
}

type LTIJWT_ToolPlatform struct {
	GUID              string `json:"g"`
	Name              string `json:"n"`
	ProductFamilyCode string `json:"p"`
	URL               string `json:"u"`
	Version           string `json:"v"`
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

type DeepLinkingTarget string

const (
	DeepLinkingTarget_Iframe DeepLinkingTarget = "iframe"
	DeepLinkingTarget_Window DeepLinkingTarget = "window"
)

type DeepLinkType string

const (
	DeepLinkType_File        DeepLinkType = "file"
	DeepLinkType_HTML        DeepLinkType = "html"
	DeepLinkType_LtiResource DeepLinkType = "ltiResourceLink"
	DeepLinkType_Image       DeepLinkType = "image"
)

type DeepLinkContext struct {
	jwt.RegisteredClaims
	Nonce            string              `json:"n"`
	ReturnAud        string              `json:"ra"`
	ReturnURL        string              `json:"r"` // deep_link_return_url
	Data             string              `json:"d"` // opaque state blob from LMS
	AcceptTypes      []DeepLinkType      `json:"a"` // allowed content types (e.g. ltiResourceLink, html, file)
	Targets          []DeepLinkingTarget `json:"t"` // allowed presentation targets (iframe, window)
	AutoCreate       bool                `json:"c"` // whether LMS auto-adds items
	AcceptMediaTypes string              `json:"m"`
	AttachedKID      string              `json:"k"` // All must be attached to a valid Session JWT
}
