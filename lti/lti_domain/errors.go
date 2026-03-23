package lti_domain

import "errors"

var (
	ErrStateNotFound                 = errors.New("state not found")
	ErrSwapTokenNotFound             = errors.New("swap token not found")
	ErrExchangeTokenNotFound         = errors.New("exchange token not found")
	ErrExchangeTokenAlreadyExchanged = errors.New("exchange token already exchanged")
	ErrExchangeRedemptionExpired     = errors.New("exchange redemption expired")
	ErrDeploymentNotFound            = errors.New("deployment not found")
)
