package server

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/pthethanh/micro/util/contextutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CorrelationIDStreamInterceptor returns a grpc.StreamServerInterceptor that provides
// a context with correlation_id for tracing. It will try to looks for value of X-Correlation-ID or X-Request-ID
// in the metadata of the incoming request. If no value is provided, a new UUID will be generated.
func CorrelationIDStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id, ok := contextutil.CorrelationIDFromContext(ss.Context())
		if ok {
			return handler(srv, ss)
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = metadata.NewIncomingContext(ss.Context(), metadata.Pairs(contextutil.XCorrelationID, id))
		return handler(srv, wrapped)
	}
}

// CorrelationIDUnaryInterceptor returns a grpc.UnaryServerInterceptor that provides
// a context with correlation_id for tracing. It will try to looks for value of X-Correlation-ID or X-Request-ID
// in the metadata of the incoming request. If no value is provided, a new UUID will be generated.
func CorrelationIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		id, ok := contextutil.CorrelationIDFromContext(ctx)
		if ok {
			return handler(ctx, req)
		}
		newCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(contextutil.XCorrelationID, id))
		return handler(newCtx, req)
	}
}
