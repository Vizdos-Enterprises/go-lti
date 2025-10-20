package lti_domain

import (
	"context"
)

const ContextKey_Session string = "lti_session"

// ContextWithLTI stores LTIJWT into the request context.
func ContextWithLTI(ctx context.Context, claims *LTIJWT) context.Context {
	return context.WithValue(ctx, ContextKey_Session, claims)
}

// LTIFromContext retrieves LTIJWT from context, if present.
func LTIFromContext(ctx context.Context) (*LTIJWT, bool) {
	val, ok := ctx.Value(ContextKey_Session).(*LTIJWT)
	return val, ok
}
