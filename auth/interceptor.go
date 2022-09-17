package auth

import (
	"context"
	"fmt"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

// HTTPInterceptor return a HTTP interceptor that perform an authentication check
// for each request using the given authenticator.
func HTTPInterceptor(a Authenticator) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wl, ok := a.(WhiteListAuthenticator); ok && wl.IsWhiteListed(r.URL.Path) {
				h.ServeHTTP(w, r)
				return
			}
			tok := tokenString(r.Context())
			if tok == "" {
				tok = r.Header.Get(AuthorizationMD)
			}
			if tok == "" {
				t, err := r.Cookie(AuthorizationMD)
				if err == nil {
					tok = t.Value
				}
			}
			md := metadata.MD{}
			if v, ok := metadata.FromIncomingContext(r.Context()); ok {
				md = v.Copy()
			}
			md.Set(AuthorizationMD, tok)
			newCtx, err := a.Authenticate(metadata.NewIncomingContext(r.Context(), md))
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"code":%d,"message":"%s"}`, codes.Unauthenticated, err), http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}

// tokenString extracts the JWT toke as a string from `ctx`.
func tokenString(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	slice, ok := md[AuthorizationMD]
	if !ok || len(slice) == 0 {
		return ""
	}
	if len(slice) > 1 {
		return ""
	}
	return slice[0]
}
