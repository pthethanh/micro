package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"google.golang.org/grpc/metadata"
)

func TestAuthMap(t *testing.T) {
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
	m := auth.MapAuthenticator{
		"":        auth1,
		"jwt":     auth1,
		"api-key": auth2,
	}
	cases := []struct {
		name string
		ctx  context.Context
		err  error
	}{
		{
			name: "default",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key1,
			})),
			err: nil,
		},
		{
			name: "correct authorization type/value pair",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "jwt " + key1,
			})),
			err: nil,
		},
		{
			name: "wrong authorization type/value pair",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "api-key " + key1,
			})),
			err: auth.ErrInvalidToken,
		},
		{
			name: "correct authorization type/value pair - another type",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "api-key " + key2,
			})),
			err: nil,
		},
		{
			name: "unknown type",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "api-key-unknown " + key1,
			})),
			err: auth.ErrInvalidToken,
		},
		{
			name: "not include authorization header",
			ctx:  metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			err:  auth.ErrAuthorizationMissing,
		},
		{
			name: "not include metadata",
			ctx:  context.Background(),
			err:  auth.ErrMetadataMissing,
		},
		{
			name: "error multiple auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Join(metadata.New(map[string]string{
				auth.AuthorizationMD: "jwt " + key1,
			}), metadata.New(map[string]string{
				auth.AuthorizationMD: "jwt " + key2,
			}))),
			err: auth.ErrMultipleAuthFound,
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
