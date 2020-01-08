package auth

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
)

// StreamInterceptor returns a grpc.StreamServerInterceptor that performs
// an authentication check for each request by using
// Authenticator.Authenticate(ctx context.Context).
func StreamInterceptor(auth Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var newCtx context.Context
		var err error
		if srvAuth, ok := srv.(Authenticator); ok {
			newCtx, err = srvAuth.Authenticate(ss.Context())
		} else {
			newCtx, err = auth.Authenticate(ss.Context())
		}
		if err != nil {
			return err
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// UnaryInterceptor returns a grpc.UnaryServerInterceptor that performs
// an authentication check for each request by using
// Authenticator.Authenticate(ctx context.Context).
func UnaryInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var newCtx context.Context
		if srvAuth, ok := info.Server.(Authenticator); ok {
			newCtx, err = srvAuth.Authenticate(ctx)
		} else {
			newCtx, err = auth.Authenticate(ctx)
		}
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}
