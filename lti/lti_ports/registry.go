package lti_ports

import (
	"context"

	"github.com/kvizdos/lti-server/lti/lti_domain"
)

type Registry interface {
	// Find a deployment
	GetDeployment(ctx context.Context, clientID string, deploymentID string) (*lti_domain.Deployment, error)

	AddDeployment(ctx context.Context, dep *lti_domain.Deployment)
}
