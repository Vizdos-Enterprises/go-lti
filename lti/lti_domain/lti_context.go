package lti_domain

import (
	"context"
)

const ContextKey_Session string = "lti_session"

const ContextKey_CookieConfirmation string = "lti_supported"

const ContextKey_SessionID string = "lti_session_id"

// ContextWithLTI stores LTIJWT into the request context.
func ContextWithLTI(ctx context.Context, claims *LTIJWT) context.Context {
	return context.WithValue(ctx, ContextKey_Session, claims)
}

// LTIFromContext retrieves LTIJWT from context, if present.
func LTIFromContext(ctx context.Context) (*LTIJWT, bool) {
	val, ok := ctx.Value(ContextKey_Session).(*LTIJWT)
	return val, ok
}
