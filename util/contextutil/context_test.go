package contextutil_test

import (
	"context"
	"testing"

	"github.com/pthethanh/micro/util/contextutil"
	"google.golang.org/grpc/metadata"
)

func TestCorrelationContext(t *testing.T) {
	expID := "123"
	// correlation id from x-correlation-id
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(contextutil.XCorrelationID, expID))
	id, ok := contextutil.CorrelationIDFromContext(ctx)
	if !ok || id != expID {
		t.Errorf("got correlation_id=%s, want correlation_id=%s", id, expID)
	}

	// correlation id from x-request-id
	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(contextutil.XRequestID, expID))
	id, ok = contextutil.CorrelationIDFromContext(ctx)
	if !ok || id != expID {
		t.Errorf("got correlation_id=%s, want correlation_id=%s", id, expID)
	}

	// generate new correlation id if not existed.
	ctx = metadata.NewIncomingContext(context.Background(), metadata.MD{})
	id, ok = contextutil.CorrelationIDFromContext(ctx)
	if ok || id == "" {
		t.Errorf("got correlation_id=%s, want correlation_id=%s", id, expID)
	}
}
