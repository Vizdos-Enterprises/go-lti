package lti_domain

import "time"

type State struct {
	Issuer       string
	ClientID     string
	DeploymentID string
	Nonce        string
	TenantID     TenantID
	CreatedAt    time.Time
}
