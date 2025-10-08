package lti_domain

import "errors"

var (
	ErrStateNotFound      = errors.New("state not found")
	ErrDeploymentNotFound = errors.New("deployment not found")
)
