package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}

func TestAuthWhiteList(t *testing.T) {
	secret := []byte("test1")
	a := jwt.Authenticator(secret)
	key, err := jwt.Encode(jwt.Claims{
		ID:        "test",
		ExpiresAt: time.Now().Add(10 * time.Second).Unix(),
	}, secret)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		ctx       context.Context
		method    string
		err       error
		matchFunc []auth.WhiteListMatchFunc
		whiteList []string
	}{
		{
			name: "whitelisted",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			matchFunc: []auth.WhiteListMatchFunc{strings.Contains},
			method:    "/helloworld.Greeter/SayHello",
			whiteList: []string{"SayHello", "DeleteHello"},
			err:       nil,
		},
		{
			name: "whitelisted regex",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			matchFunc: []auth.WhiteListMatchFunc{auth.WhiteListMatchFuncRegexp},
			method:    "/helloworld.Greeter/SayHello",
			whiteList: []string{".*Say.*", "DeleteHello"},
			err:       nil,
		},
		{
			name: "not whitelisted and not valid auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			matchFunc: []auth.WhiteListMatchFunc{strings.Contains},
			method:    "/helloworld.Greeter/SaySomething",
			whiteList: []string{"SayHello", "DeleteHello"},
			err:       auth.ErrInvalidToken,
		},
		{
			name: "not whitelisted but valid auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key,
			})),
			method:    "/helloworld.Greeter/SaySomething",
			whiteList: []string{"SayHello", "DeleteHello"},
			matchFunc: []auth.WhiteListMatchFunc{},
			err:       nil,
		},
	}

	for _, c := range cases {
		wa := auth.NewWhiteListAuthenticator(a, c.whiteList, c.matchFunc...)
		t.Run(c.name, func(t *testing.T) {
			// unary
			_, err = auth.UnaryInterceptor(wa)(c.ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: c.method,
			}, grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			}))
			if err != c.err {
				t.Errorf("unary, got err=%v, want err=%v", err, c.err)
			}
			// stream
			err = auth.StreamInterceptor(wa)(nil, &serverStream{ctx: c.ctx}, &grpc.StreamServerInfo{
				FullMethod: c.method,
			}, grpc.StreamHandler(func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}))
			if err != c.err {
				t.Errorf("stream, got err=%v, want err=%v", err, c.err)
			}
		})
	}

}
