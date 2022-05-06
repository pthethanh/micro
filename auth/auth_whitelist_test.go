package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		name   string
		ctx    context.Context
		method string
		err    error
		funcs  []auth.WhiteListFunc
	}{
		{
			name: "whitelisted",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/helloworld.Greeter/SayHello")},
			method: "/helloworld.Greeter/SayHello",
			err:    nil,
		},
		{
			name: "whitelisted regex",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListRegexp(".*Say.*")},
			method: "/helloworld.Greeter/SayHello",
			err:    nil,
		},
		{
			name: "whitelisted regex, not match",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListRegexp(".*Speak.*")},
			method: "/helloworld.Greeter/SayHello",
			err:    auth.ErrInvalidToken,
		},
		{
			name: "not whitelisted and not valid auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/helloworld.Greeter/SaySomethingElse")},
			method: "/helloworld.Greeter/SaySomething",
			err:    auth.ErrInvalidToken,
		},
		{
			name: "not whitelisted but valid auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key,
			})),
			method: "/helloworld.Greeter/SaySomething",
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/helloworld.Greeter/SaySomethingElse")},
			err:    nil,
		},
		{
			name: "whitelist using NOT",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			method: "/helloworld.Greeter/SaySomething",
			funcs:  []auth.WhiteListFunc{auth.WhiteListNot(auth.WhiteListInList("/helloworld.Greeter/SaySomethingElse"))},
			err:    nil,
		},
		{
			name: "whitelist using NOT, invalid key",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			method: "/helloworld.Greeter/SaySomething",
			funcs:  []auth.WhiteListFunc{auth.WhiteListNot(auth.WhiteListInList("/helloworld.Greeter/SaySomething"))},
			err:    auth.ErrInvalidToken,
		},
	}

	for _, c := range cases {
		wa := auth.NewWhiteListAuthenticator(a, c.funcs...)
		t.Run(c.name, func(t *testing.T) {
			// unary
			_, err = auth.UnaryInterceptor(wa)(c.ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: c.method,
			}, grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
				return nil, nil
			}))
			if err != c.err {
				t.Errorf("unary, got err=%v, want err=%v", err, c.err)
			}
			// stream
			err = auth.StreamInterceptor(wa)(nil, &serverStream{ctx: c.ctx}, &grpc.StreamServerInfo{
				FullMethod: c.method,
			}, grpc.StreamHandler(func(srv any, stream grpc.ServerStream) error {
				return nil
			}))
			if err != c.err {
				t.Errorf("stream, got err=%v, want err=%v", err, c.err)
			}
		})
	}

}

func TestAuthWhiteListHTTP(t *testing.T) {
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
		name    string
		ctx     context.Context
		path    string
		status  int
		funcs   []auth.WhiteListFunc
		headers map[string]string
		cookie  *http.Cookie
	}{
		{
			name: "whitelisted",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v1/test")},
			path:   "/api/v1/test",
			status: http.StatusOK,
		},
		{
			name:   "not whitelisted, no metadata",
			ctx:    context.Background(),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			path:   "/api/v1/test",
			status: http.StatusUnauthorized,
		},
		{
			name:   "not whitelisted, invalid metadata - empty",
			ctx:    metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			path:   "/api/v1/test",
			status: http.StatusUnauthorized,
		},
		{
			name:   "not whitelisted, invalid metadata - more than 1 meta",
			ctx:    metadata.NewIncomingContext(context.Background(), metadata.Pairs(auth.AuthorizationMD, key, auth.AuthorizationMD, key)),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			path:   "/api/v1/test",
			status: http.StatusUnauthorized,
		},
		{
			name: "whitelisted regex",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListRegexp("/api/.*")},
			path:   "/api/v1/test",
			status: http.StatusOK,
		},
		{
			name: "whitelisted regex, not match",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListRegexp("/api/v2/.*")},
			path:   "/api/v1/test",
			status: http.StatusUnauthorized,
		},
		{
			name: "not whitelisted and not valid auth",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			path:   "/api/v1/test",
			status: http.StatusUnauthorized,
		},
		{
			name: "not whitelisted but valid auth, use metadata",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: key,
			})),
			path:   "/api/v1/test",
			funcs:  []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			status: http.StatusOK,
		},
		{
			name:    "not whitelisted but valid auth, use header",
			path:    "/api/v1/test",
			funcs:   []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			headers: map[string]string{auth.AuthorizationMD: key},
			status:  http.StatusOK,
		},
		{
			name:  "not whitelisted but valid auth, use cookie",
			path:  "/api/v1/test",
			funcs: []auth.WhiteListFunc{auth.WhiteListInList("/api/v2/test")},
			cookie: &http.Cookie{
				Name:  auth.AuthorizationMD,
				Value: key,
			},
			status: http.StatusOK,
		},
		{
			name: "whitelist using NOT",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			path:   "/api/v1/test",
			funcs:  []auth.WhiteListFunc{auth.WhiteListNot(auth.WhiteListInList("/api/v2/test"))},
			status: http.StatusOK,
		},
		{
			name: "whitelist using NOT, invalid key",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				auth.AuthorizationMD: "invalid",
			})),
			path:   "/api/v1/test",
			funcs:  []auth.WhiteListFunc{auth.WhiteListNot(auth.WhiteListInList("/api/v1/test"))},
			status: http.StatusUnauthorized,
		},
	}

	for _, c := range cases {
		wa := auth.NewWhiteListAuthenticator(a, c.funcs...)
		t.Run(c.name, func(t *testing.T) {
			host := "http://localhost"
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, host+c.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			if c.ctx != nil {
				req = req.WithContext(c.ctx)
			}
			if c.headers != nil {
				for k, v := range c.headers {
					req.Header.Set(k, v)
				}
			}
			if c.cookie != nil {
				req.AddCookie(c.cookie)
			}
			auth.HTTPInterceptor(wa)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			})).ServeHTTP(recorder, req)
			gotStatus := recorder.Result().StatusCode
			if gotStatus != c.status {
				t.Errorf("got status_code=%d, want status_code=%d", gotStatus, c.status)
			}
		})
	}

}
