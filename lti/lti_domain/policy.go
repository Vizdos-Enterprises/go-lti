package lti_domain

import "time"

type Policy struct {
	MaxClockSkew time.Duration
	StateTTL     time.Duration
	NonceTTL     time.Duration
	AllowedURIs  []string
}
