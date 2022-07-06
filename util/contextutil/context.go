package contextutil

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const (
	// XCorrelationID correlation id header.
	XCorrelationID = "x-correlation-id"
	// XRequestID request id header.
	XRequestID = "x-request-id"
)

// CorrelationIDFromContext tries to get value of X-Correlation-ID then X-Request-ID from meta data.
// If no value is provided, a new UUID value will be return.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	// assume it's the incoming context.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md[XCorrelationID]; ok {
			return v[0], true
		}
		if v, ok := md[XRequestID]; ok {
			return v[0], true
		}
		return correlationIDFromNormalContext(ctx)
	}
	// if not, try to get from outgoing context
	md, ok = metadata.FromOutgoingContext(ctx)
	if ok {
		if v, ok := md[XCorrelationID]; ok {
			return v[0], true
		}
		if v, ok := md[XRequestID]; ok {
			return v[0], true
		}
		return correlationIDFromNormalContext(ctx)
	}
	// otherwise, it's probably a normal context.
	return correlationIDFromNormalContext(ctx)
}

func correlationIDFromNormalContext(ctx context.Context) (string, bool) {
	if id := ctx.Value(XCorrelationID); id != nil && fmt.Sprintf("%v", id) != "" {
		return fmt.Sprintf("%v", id), true
	}
	if id := ctx.Value(XRequestID); id != nil && fmt.Sprintf("%v", id) != "" {
		return fmt.Sprintf("%v", id), true
	}
	// if no id in context, return new generated id.
	return uuid.New().String(), false
}
