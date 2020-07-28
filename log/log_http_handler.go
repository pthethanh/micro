package log

import (
	"net/http"

	"github.com/google/uuid"
)

// NewHTTPContextHandler adds a context logger based on the given logger to
// each request. After a request passes through this handler,
// WithContext(req.Context()).Error(, "foo") will log to that logger and add useful context
// to each log entry.
func NewHTTPContextHandler(l Logger) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// allow requests in microservices environment can be traced.
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = r.Header.Get("X-Request-ID")
			}
			if correlationID == "" {
				correlationID = uuid.New().String()
			}
			logger := l.Fields(
				"correlation_id", correlationID,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"method", r.Method)
			r = r.WithContext(NewContext(ctx, logger))
			inner.ServeHTTP(w, r)
		})
	}
}
