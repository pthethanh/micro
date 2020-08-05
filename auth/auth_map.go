package auth

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// MapAuthenticator chains multiple Authenticators,
// allowing multiple authentication types to be used with a single
// interceptor. Key of the map wil be used to match with authorization type in metadata.
type MapAuthenticator map[string]Authenticator

// Authenticate implements the Authenticator interface.
func (m MapAuthenticator) Authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetadataMissing
	}
	slice, ok := md[AuthorizationMD]
	if !ok || len(slice) == 0 {
		return nil, ErrAuthorizationMissing
	}
	if len(slice) > 1 {
		return nil, ErrMultipleAuthFound
	}
	kv := strings.Fields(slice[0])
	k := ""
	v := ""
	if len(kv) == 1 {
		v = kv[0]
	} else if len(kv) == 2 {
		k = kv[0]
		v = kv[1]
	}
	if a, ok := m[k]; ok {
		md := md.Copy()
		md.Set(AuthorizationMD, v)
		return a.Authenticate(metadata.NewIncomingContext(ctx, md))
	}
	return nil, ErrInvalidToken
}
