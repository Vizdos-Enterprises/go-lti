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

type Deployment interface {
	GetDeploymentID() string
	GetTenantID() TenantID
	GetLTIIssuer() string
	GetLTIClientID() string
	GetLTIJWKSURL() string
	GetLTIAuthEndpoint() string
	GetLTITokenEndpoint() string
	GetLTIDeploymentID() string
}

var _ Deployment = (*BaseLTIDeployment)(nil)
var _ Deployment = BaseLTIDeployment{}

type BaseLTIDeployment struct {
	InternalID    string
	ForTenantID   string
	Issuer        string
	ClientID      string
	JWKSURL       string
	AuthEndpoint  string
	TokenEndpoint string
	DeploymentID  string
}

func (d BaseLTIDeployment) GetDeploymentID() string {
	return d.DeploymentID
}

func (d BaseLTIDeployment) GetTenantID() TenantID {
	return d.ForTenantID
}

func (d BaseLTIDeployment) GetLTIIssuer() string {
	return d.Issuer
}

func (d BaseLTIDeployment) GetLTIClientID() string {
	return d.ClientID
}

func (d BaseLTIDeployment) GetLTIJWKSURL() string {
	return d.JWKSURL
}

func (d BaseLTIDeployment) GetLTIAuthEndpoint() string {
	return d.AuthEndpoint
}

func (d BaseLTIDeployment) GetLTITokenEndpoint() string {
	return d.TokenEndpoint
}

func (d BaseLTIDeployment) GetLTIDeploymentID() string {
	return d.DeploymentID
}
