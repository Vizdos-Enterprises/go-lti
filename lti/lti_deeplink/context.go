package lti_deeplink

import (
	"context"

	"github.com/vizdos-enterprises/go-lti/internal/adapters/deeplinking"
	"github.com/vizdos-enterprises/go-lti/lti/lti_domain"
)

// ContextWithLTI stores LTIJWT into the request context.
func ContextWithDeepLink(ctx context.Context, claims *lti_domain.DeepLinkContext) context.Context {
	return context.WithValue(ctx, deeplinking.ContextKey_DeepLink, claims)
}

func DeepLinkFromContext(ctx context.Context) (*lti_domain.DeepLinkContext, bool) {
	val, ok := ctx.Value(deeplinking.ContextKey_DeepLink).(*lti_domain.DeepLinkContext)
	return val, ok
}
