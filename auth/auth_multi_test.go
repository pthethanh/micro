package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"google.golang.org/grpc/metadata"
)

func TestAuthMulti(t *testing.T) {
	s1 := []byte("test1")
	auth1 := jwt.Authenticator(s1)
	key1, err := jwt.Encode(jwt.Claims{
		StandardClaims: jwt.StandardClaims{
			Id:        "test",
			ExpiresAt: time.Now().Add(10 * time.Second).Unix(),
		},
	}, s1)
	if err != nil {
		t.Fatal(err)
	}
	s2 := []byte("test2")
	key2, err := jwt.Encode(jwt.Claims{}, s2)
	if err != nil {
		t.Fatal(err)
	}
	auth2 := jwt.Authenticator(s2)
	m := auth.MultiAuthenticator{auth1, auth2}
	cases := []struct {
		name string
		ctx  context.Context
		err  error
	}{
		{
			name: "auth with first authenticator in the list",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key1,
			})),
			err: nil,
		},
		{
			name: "auth with second authenticator in the list",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key2,
			})),
			err: nil,
		},
		{
			name: "not include authorization header",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			err:  auth.ErrAuthorizationMissing,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := m.Authenticate(c.ctx); err != c.err {
				t.Errorf("got err=%v, want err=%v", err, c.err)
			}
		})
	}
}
