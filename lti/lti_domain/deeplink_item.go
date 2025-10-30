package lti_domain

type DeepLinkItem struct {
	Type    DeepLinkType        `json:"type"`
	Title   string              `json:"title"`
	URL     string              `json:"url"`
	Custom  map[string]string   `json:"custom,omitempty"`
	Targets []DeepLinkingTarget `json:"presentation_document_target,omitempty"`

	LineItem *DeepLinkLineItem `json:"lineItem,omitempty"`
}

type DeepLinkLineItem struct {
	Label         string  `json:"label,omitempty"`
	ScoreMaximum  float64 `json:"scoreMaximum,omitempty"`
	ResourceID    string  `json:"resourceId,omitempty"`
	Tag           string  `json:"tag,omitempty"`
	StartDateTime string  `json:"startDateTime,omitempty"`
	EndDateTime   string  `json:"endDateTime,omitempty"`
}
