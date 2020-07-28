package log

import (
	"net/http"

	"github.com/google/uuid"
)

// NewHTTPContextHandler provides a context logger with correlation_id and other basic HTTP information.
// Value of correlation_id will be retrieved from X-Correlation-ID or X-Request-ID in header of the incoming request.
// If no value of correlation_id is provided, a new UUID will be generated.
func NewHTTPContextHandler(l Logger) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// allow requests in microservices environment can be traced.
			logger := l.Fields(
				correlationIDKey, getCorrelationID(r),
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"method", r.Method)
			r = r.WithContext(NewContext(ctx, logger))
			inner.ServeHTTP(w, r)
		})
	}
}

func getCorrelationID(r *http.Request) string {
	if id := r.Header.Get("X-Correlation-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}
