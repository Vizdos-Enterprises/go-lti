package lti_domain

import "time"

type LaunchMethod uint8

const (
	LaunchMethodUnknown LaunchMethod = iota
	LaunchMethodDirect
	LaunchMethodPKCE
)

func (m LaunchMethod) String() string {
	switch m {
	case LaunchMethodDirect:
		return "DirectLaunch"
	case LaunchMethodPKCE:
		return "PKCELaunch"
	default:
		return "Unknown"
	}
}

type LaunchEvent struct {
	At        time.Time
	Method    LaunchMethod
	Success   bool
	Platform  string
	UserAgent string

	Duration time.Duration
}
