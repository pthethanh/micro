package log

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NewHTTPContextHandler provides a context logger with correlation_id and other basic HTTP information.
// Value of correlation_id will be retrieved from X-Correlation-ID or X-Request-ID in header of the incoming request.
// If no value of correlation_id is provided, a new UUID will be generated.
// This middleware should be used for HTTP handler only.
// Since this method use a Hijack response writer, it cannot be used with option server.HTTPInterceptors.
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
			mw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			bg := time.Now()
			logger.Fields("phase", "request").Info("request started")
			inner.ServeHTTP(mw, r)
			logger.Fields("phase", "response", "duration", time.Since(bg), "status_code", mw.status).Info("request finished")
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

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.status = code
}

func (w *responseWriter) Status() int {
	return w.status
}
