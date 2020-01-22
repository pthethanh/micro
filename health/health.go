package health

import (
	"context"
	"net/http"
	"sync"
)

var (
	isReadyMu sync.RWMutex
	isReady   bool
)

// CheckFunc function signature for health checks.
type CheckFunc func(context.Context) error

// Ready marks the service as ready to receive traffic.
func Ready() {
	isReadyMu.Lock()
	isReady = true
	isReadyMu.Unlock()
}

// Readiness returns an HTTP handler for checking Readiness state.
// Will return 503 until Ready() is called
func Readiness() http.Handler {
	ok := []byte("OK")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isReadyMu.RLock()
		defer isReadyMu.RUnlock()
		if !isReady {
			http.Error(w, "Not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write(ok)
	})
}

// Liveness returns an HTTP handler for checking the health of the service.
func Liveness(cf ...CheckFunc) http.Handler {
	ok := []byte("OK")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for _, c := range cf {
			if err := c(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		w.Write(ok)
	})
}
