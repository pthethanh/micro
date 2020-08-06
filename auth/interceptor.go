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
		a := auth
		// if server overrides Authenticator, use it instead.
		if srvAuth, ok := srv.(Authenticator); ok {
			a = srvAuth
		}
		if wl, ok := a.(WhiteListAuthenticator); ok && wl.IsWhiteListed(info.FullMethod) {
			newCtx, err = ss.Context(), nil
		} else {
			newCtx, err = a.Authenticate(ss.Context())
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
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error
		a := auth
		// if server override Authenticator, use it instead.
		if srvAuth, ok := info.Server.(Authenticator); ok {
			a = srvAuth
		}
		if wl, ok := a.(WhiteListAuthenticator); ok && wl.IsWhiteListed(info.FullMethod) {
			newCtx, err = ctx, nil
		} else {
			newCtx, err = a.Authenticate(ctx)
		}
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}
