package observability

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func CaptureRequestError(r *http.Request, err error, msg string, kv ...any) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(r.Context())
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)

	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for i := 0; i+1 < len(kv); i += 2 {
				key, _ := kv[i].(string)
				scope.SetTag(key, toString(kv[i+1]))
			}
			if traceID, ok := r.Context().Value("trace_id").(string); ok {
				scope.SetTag("trace_id", traceID)
			}
			if requestID, ok := r.Context().Value("request_id").(string); ok {
				scope.SetTag("request_id", requestID)
			}
			hub.CaptureException(err)
		})
		return
	}

	sentry.CaptureException(err)
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		return ""
	}
}
