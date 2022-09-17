package log

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pthethanh/micro/util/contextutil"

	"google.golang.org/grpc"
)

// StreamInterceptor returns a grpc.StreamServerInterceptor that provides
// a context logger with correlation_id. It will try to looks for value of X-Correlation-ID or X-Request-ID
// in the metadata of the incoming request. If no value is provided, a new UUID will be generated.
// For REST API via gRPC Gateway, pass the value of X-Correlation-ID or X-Request-ID in the header.
func StreamInterceptor(l Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		correlationID, _ := contextutil.CorrelationIDFromContext(ss.Context())
		logger := l.Fields(CorrelationID, correlationID, "method", info.FullMethod)
		newCtx := NewContext(ss.Context(), logger)
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		_, err := runWithLog(logger, info.FullMethod, func() (interface{}, error) {
			err := handler(srv, wrapped)
			return nil, err
		})
		return err
	}
}

// UnaryInterceptor returns a grpc.UnaryServerInterceptor that provides
// a context logger with correlation_id. It will try to looks for value of X-Correlation-ID or X-Request-ID
// in the metadata of the incoming request. If no value is provided, a new UUID will be generated.
// For REST API via gRPC Gateway, pass the value of X-Correlation-ID or X-Request-ID in the header.
func UnaryInterceptor(l Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		correlationID, _ := contextutil.CorrelationIDFromContext(ctx)
		logger := l.Fields(CorrelationID, correlationID, "method", info.FullMethod)
		newCtx := NewContext(ctx, logger)
		return runWithLog(logger, info.FullMethod, func() (interface{}, error) {
			return handler(newCtx, req)
		})
	}
}

func runWithLog(logger Logger, method string, f func() (interface{}, error)) (interface{}, error) {
	bg := time.Now()
	logger.Fields("phase", "request").Info("request started")
	res, err := f()
	status := "success"
	if err != nil {
		status = "failed"
	}
	logger.Fields("phase", "response", "duration", time.Since(bg), "status", status, "error", err).Info("request finished")
	return res, err
}
