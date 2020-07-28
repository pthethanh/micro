package log

import (
	"context"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// StreamInterceptor returns a grpc.StreamServerInterceptor that provides
// a context logger with correlation-id. It will try to looks for value of X-Correlation-ID or X-Request-ID
// in the metadata of the incoming request. If no value is provided, a new one will be generated.
// For REST API, use Grpc-Metadata-Request-Id as header key for passing a request id.
func StreamInterceptor(l Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		correlationID := correlationIDFromGRPCContext(ss.Context())
		logger := l.Fields("correlation_id", correlationID)
		newCtx := NewContext(ss.Context(), logger)
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// UnaryInterceptor returns a grpc.UnaryServerInterceptor that provides
// a context logger with request_id. If a request-id is already specified
// in the metadata, it will be used. Otherwise a new one will be generated.
// For REST API, use Grpc-Metadata-Request-Id as header key for passing a request id.
func UnaryInterceptor(l Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		correlationID := correlationIDFromGRPCContext(ctx)
		logger := l.Fields("correlation_id", correlationID)
		newCtx := NewContext(ctx, logger)
		return handler(newCtx, req)
	}
}

// try to get from meta data from X-Correlation-ID then X-Request-ID.
// otherwise generate a new one.
func correlationIDFromGRPCContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md["x-correlation-id"]; ok {
			return v[0]
		}
		if v, ok := md["x-request-id"]; ok {
			return v[0]
		}
	}
	return uuid.New().String()
}
