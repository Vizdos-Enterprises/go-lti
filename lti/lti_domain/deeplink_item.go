package lti_domain

type DeepLinkItem struct {
	Type    DeepLinkType        `json:"type"`
	Title   string              `json:"title"`
	URL     string              `json:"url"`
	Custom  map[string]string   `json:"custom,omitempty"`
	Targets []DeepLinkingTarget `json:"presentation_document_target,omitempty"`
}
