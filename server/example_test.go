package server_test

import (
	"context"
	"net/http"

	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"
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
		server.AuthJWT("secret"),
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
		server.HTTPHandler("/doc", h),
	)
	if err := srv.ListenAndServe( /*services ...Service*/ ); err != nil {
		panic(err)
	}
}
