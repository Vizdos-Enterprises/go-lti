package lti_domain

import (
	"context"
)

type ctxKey int

const ltiContextKey ctxKey = iota

// ContextWithLTI stores LTIJWT into the request context.
func ContextWithLTI(ctx context.Context, claims *LTIJWT) context.Context {
	return context.WithValue(ctx, ltiContextKey, claims)
}

// LTIFromContext retrieves LTIJWT from context, if present.
func LTIFromContext(ctx context.Context) (*LTIJWT, bool) {
	val, ok := ctx.Value(ltiContextKey).(*LTIJWT)
	return val, ok
}
