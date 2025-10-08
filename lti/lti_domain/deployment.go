package lti_domain

import "fmt"

type InternalDeploymentID any

func DeploymentIDToString(id InternalDeploymentID) string {
	switch v := id.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

type Deployment struct {
	InternalID    InternalDeploymentID
	ForTenantID   TenantID
	Issuer        string
	ClientID      string
	JWKSURL       string
	AuthEndpoint  string
	TokenEndpoint string
	DeploymentID  string
}
