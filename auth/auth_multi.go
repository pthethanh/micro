package auth

import "golang.org/x/net/context"

// MultiAuthenticator chains a series of Authenticators,
// allowing multiple authentication types to be used with a single
// interceptor.
type MultiAuthenticator []Authenticator

// Authenticate implements the Authenticator interface.
func (m MultiAuthenticator) Authenticate(ctx context.Context) (context.Context, error) {
	var err error
	var newCtx context.Context
	for _, a := range m {
		newCtx, err = a.Authenticate(ctx)
		if err == nil {
			return newCtx, nil
		}
	}
	return nil, err
}
