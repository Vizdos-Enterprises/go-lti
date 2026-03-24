package lti_domain

import (
	"time"
)

type State struct {
	Issuer       string
	ClientID     string
	DeploymentID string
	Nonce        string
	TenantID     TenantID
	CreatedAt    time.Time
}

type SwapToken struct {
	To          string    `json:"to"`
	RequestorUA string    `json:"ua"`
	Claims      LTIJWT    `json:"jwt"`
	StartAt     time.Time `json:"sa"`
}

type ExchangeToken struct {
	Data           *SwapToken `json:"d"`
	ClaimableUntil time.Time  `json:"u"`
	Exchanged      bool       `json:"e"`
	Challenge      string     `json:"c,omitempty"`
	AuthToken      string     `json:"t,omitempty"`
}
