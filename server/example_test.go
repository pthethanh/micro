package server_test

import (
	"context"
	"net/http"

	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

func ExampleListenAndServe() {
	if err := server.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}

func ExampleListenAndServeContext() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.ListenAndServeContext(ctx /*, services ...Service*/); err != nil {
		panic(err)
	}
}

func ExampleNew_fromEnvironmentVariables() {
	srv := server.New(server.FromEnv())
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		log.Panic(err)
	}
}

func ExampleNew_withOptions() {
	srv := server.New(
		server.Address(":8080"),
		server.JWT("secret"),
		server.Logger(log.Fields("service", "micro")),
	)
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}

func ExampleNew_withSinglePageApplication() {
	// See https://github.com/pthethanh/micro/tree/master/examples/helloworld/web for a full runnable example.
	srv := server.New(
		server.Address(":8080"),
		// routes all calls to /api/ to gRPC Gateway handlers.
		server.APIPrefix("/api/"),
		// serve web at /
		server.Web("/", "public", "index.html"),
	)
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}

func ExampleNew_withInternalHTTPAPI() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("doc"))
	})
	srv := server.New(
		server.FromEnv(),
		server.Handler("/doc", h),
	)
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}

func ExampleNew_withExternalInterceptors() {
	srv := server.New(
		server.FromEnv(),
		server.StreamInterceptors(
			tags.StreamServerInterceptor(),
			tracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			recovery.StreamServerInterceptor(),
		),
		server.UnaryInterceptors(
			tags.UnaryServerInterceptor(),
			tracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(),
		),
	)
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}
