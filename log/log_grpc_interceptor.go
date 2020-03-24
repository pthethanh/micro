package log

import (
	"context"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// StreamInterceptor returns a grpc.StreamServerInterceptor that provides
// a context logger with request_id. If a request-id is already specified
// in the metadata, it will be used. Otherwise a new one will be generated.
// For REST API, use Grpc-Metadata-Request-Id as header key for passing a request id.
func StreamInterceptor(l Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		requestID := requestIDFromGRPCContext(ss.Context())
		logger := l.WithFields(Fields{
			"request_id": requestID,
		})
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
		requestID := requestIDFromGRPCContext(ctx)
		logger := l.WithFields(Fields{
			"request_id": requestID,
		})
		newCtx := NewContext(ctx, logger)
		return handler(newCtx, req)
	}
}

// try to get from meta data first.
// otherwise generate a new one.
func requestIDFromGRPCContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md["request-id"]; ok {
			return v[0]
		}
	}
	return uuid.New().String()
}
