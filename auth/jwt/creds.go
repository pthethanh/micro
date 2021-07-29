package jwt

import (
	"golang.org/x/net/context"

	"google.golang.org/grpc/credentials"
)

// NewJWTCredentialsFromToken returns a grpc rpc credential
// using the provided JWT token. Does not validate the Token.
func NewJWTCredentialsFromToken(token string) credentials.PerRPCCredentials {
	return jwtToken(token)
}

type jwtToken string

// GetRequestMetadata implements the `credentials.PerRPCCredentials`
// method `GetRequestMetadata`, by returning a simple map to be appended
// to GRPC metadata.
func (j jwtToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": string(j),
	}, nil
}

// RequireTransportSecurity implements `credentials.PerRPCCredentials`
// method `RequireTransportSecurity`. Indicates if we want a secure
// transport.
func (j jwtToken) RequireTransportSecurity() bool {
	return false
}
