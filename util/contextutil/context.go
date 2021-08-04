package contextutil

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const (
	XCorrelationID = "x-correlation-id"
	XRequestID     = "x-request-id"
)

// CorrelationIDFromContext tries to get value of X-Correlation-ID then X-Request-ID from meta data.
// If no value is provided, a new UUID value will be return.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md[XCorrelationID]; ok {
			return v[0], true
		}
		if v, ok := md[XRequestID]; ok {
			return v[0], true
		}
	}
	return uuid.New().String(), false
}
