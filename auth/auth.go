// Package auth defines standard interface for authentication.
package auth

import "context"

const (
	// AuthorizationMD authorization metadata name.
	AuthorizationMD = "authorization"

	// GrpcGWCookieMD is cookie metadata name of GRPC in gRPC GateWay Request.
	GrpcGWCookieMD = "grpcgateway-cookie"
)

// Authenticator defines the interface to perform the actual
// authentication of the request. Implementations should fetch
// the required data from the context.Context object. GRPC specific
// data like `metadata` and `peer` is available on the context.
// Should return a new `context.Context` that is a child of `ctx`
// or `codes.Unauthenticated` when auth is lacking or
// `codes.PermissionDenied` when lacking permissions.
type Authenticator interface {
	Authenticate(ctx context.Context) (context.Context, error)
}

// AuthenticatorFunc defines a pluggable function to perform authentication
// of requests. Should return a new `context.Context` that is a
// child of `ctx` or `codes.Unauthenticated` when auth is lacking or
// `codes.PermissionDenied` when lacking permissions.
type AuthenticatorFunc func(ctx context.Context) (context.Context, error)

// Authenticate implements the Authenticator interface
func (f AuthenticatorFunc) Authenticate(ctx context.Context) (context.Context, error) {
	return f(ctx)
}
