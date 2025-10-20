package lti_domain

type LTIService string

const (
	LTIService_ResourceLink LTIService = "LtiResourceLinkRequest"
	LTIService_DeepLink     LTIService = "LtiDeepLinkingRequest"
)
