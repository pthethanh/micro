# micro
[![Actions Status](https://github.com/pthethanh/micro/workflows/Go/badge.svg)](https://github.com/pthethanh/micro/actions)
[![GoDoc](https://godoc.org/github.com/pthethanh/micro?status.svg)](https://pkg.go.dev/mod/github.com/pthethanh/micro)
[![GoReportCard](https://goreportcard.com/badge/github.com/pthethanh/micro)](https://goreportcard.com/report/github.com/pthethanh/micro)

Just a simple tool kit for building microservices.

## What is micro?

micro is a Go tool kit for enterprise targeted for microservices or well designed monolith application. It doesn't aim to be a framework, but just a microservices tool kit/library for easily and quickly build API applications.

micro's vision is to be come a good tool kit for beginner/intermediate developers and hence it should be:

- Easy to use.
- Compatible with Go, gRPC native libraries.
- Come with default ready to use features.
- Backward compatible.

I expect micro requires no more than 15 minutes for a beginner/intermediate developer to be able to use the tool kit effectively. This means micro will come with lots of useful default features, but at the same time provide developers ability to provide their alternatives.

micro is built around gRPC. It exposes both gRPC and REST API over 1 single port using [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway), with default ready to use logger, metrics, health check APIs.

Currently micro comes with a collection of plugins that can be found [here](https://github.com/pthethanh/micro/tree/master/plugins)

## Getting Started

### Start your own

Define your proto, an example can be found [here](https://github.com/pthethanh/micro/blob/master/examples/helloworld/helloworld/helloworld.proto):

```proto
syntax = "proto3";

package helloworld;
option go_package = "github.com/pthethanh/micro/examples/helloworld/helloworld;helloworld";

import "google/api/annotations.proto";

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (google.api.http) = {
      post: "/api/v1/hello"
      body: "*"
    };
  }
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
```

Generate boiler plate code with [protoc-gen-micro](https://github.com/pthethanh/micro/blob/master/cmd/protoc-gen-micro), an example can be found [here](https://github.com/pthethanh/micro/blob/master/Makefile#L63):
```shell
$(PROTOC_ENV) protoc -I $(PROTOC_INCLUDES) -I $(GOOGLE_APIS_PROTO) -I ./examples/helloworld/helloworld \
	 --go_out $(PROTO_OUT) \
	 --micro_out $(PROTO_OUT) \
	 --micro_opt generate_gateway=true \
	 --go-grpc_out $(PROTO_OUT) \
	 --grpc-gateway_out $(PROTO_OUT) \
     --grpc-gateway_opt logtostderr=true \
     --grpc-gateway_opt generate_unbound_methods=true \
     ./examples/helloworld/helloworld/helloworld.proto
```

Create new gRPC service, an example can be found [here](https://github.com/pthethanh/micro/blob/master/examples/helloworld/server/main.go)

```go
type (
	service struct {
        // Embed the unimplemented server so that all the registration methods are available to the micro server.
		pb.UnimplementedGreeterServer
	}
)

// SayHello implements pb.GreeterServer interface.
func (s *service) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Context(ctx).Info("name", req.Name)
	if req.Name == "" {
		return nil, status.InvalidArgument("name must not be empty")
	}
	return &pb.HelloReply{
		Message: "Hello " + req.GetName(),
	}, nil
}
```

Start a simple server, get configurations from environment variables.

```go
package main

import (
    "github.com/pthethanh/micro/server"
)

func main() {
    srv := &service{}
    if err := server.ListenAndServe(srv); err != nil {
        panic(err)
    }
}
```

More complex with custom options.

```go
package main

import (
    "github.com/pthethanh/micro/log"
    "github.com/pthethanh/micro/server"
)

func main() {
    srv := server.New(
        server.FromEnv(),
        server.PProf(""),
        server.Address(":8088"),
        server.JWT("secret"),
        server.Web("/", "web", "index.html"),
        server.Logger(log.Fields("service", "my_service")),
        server.CORS(true, []string{"*"}, []string{"POST"}, []string{"http://localhost:8080"}),
    )
    if err := srv.ListenAndServe( /*services...*/ ); err != nil {
        panic(err)
    }
}

```

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/server?tab=doc) for more options.

## Features

Currently, micro supports following features:

### Server

- Exposes both gRPC and REST in 1 single port.
- Internal APIs:
  - Prometheus metrics.
  - Health checks.
  - Debug profiling.
- Context logging/tracing with X-Request-Id/X-Correlation-Id header/metadata.
- Authentication interceptors
- Other options: CORS, HTTP Handler, Serving Single Page Application, Interceptors,...

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/server?tab=doc) and [examples](https://pkg.go.dev/github.com/pthethanh/micro/server?tab=doc#pkg-examples) for more detail.

### Auth

- Authenticator interface.
- JWT
- Authenticator, WhiteList, Chains.
- Interceptors for both gRPC & HTTP

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/auth?tab=doc) for  more detail.

### Broker

- Standard message broker interface.
- Memory broker.
- NATS plugin.
- More plugins can be found [here](https://github.com/pthethanh/micro/tree/master/plugins/broker).

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/broker?tab=doc) for  more detail.

### Cache

- Standard cache service interface.
- Memory cache.
- Redis plugin.
- More plugins can be found [here](https://github.com/pthethanh/micro/tree/master/plugins/cache).

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/cache?tab=doc) for  more detail.

### Config

- Standard config interface.
- Config from environment variables.
- Config from file and other options.

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/config?tab=doc) for  more detail.

### Health

- Health check for readiness and liveness.
- Utilities for checking health.

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/health?tab=doc) for  more detail.

### Log

- Standard logger interface.
- Logrus implementation.
- Context logger & tracing using X-Request-Id and X-Correlation-Id
- Interceptors for HTTP & gRPC.

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/log?tab=doc) for  more detail.

### Util

- Some utilities that might need during the development using micro.

See [doc](https://pkg.go.dev/github.com/pthethanh/micro/util?tab=doc) for  more detail.

### Interceptors and Other Options

micro is completely compatible with Go native and gRPC native, hence you can use external interceptors and other external libraries along with the provided options.

Interceptors: [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)

See [examples](https://pkg.go.dev/github.com/pthethanh/micro/server?tab=doc#example_New_withExternalInterceptors) for more detail.

## Why a new standard libraries?

micro is inspired by [go-kit](https://github.com/go-kit/kit) and [go-micro](https://github.com/micro/go-micro).

go-kit is a good tool kit, but one of the thing I don't like go-kit is its over use of interface{} which cause a lot of unnecessary type conversions and some of other abstractions in the libraries which are not compatible with Go native libraries. Although go-kit is very flexible, it's a little bit hard to use for beginner/intermediate developers. It has a lot of options for developers to choose and hence hard to force everyone inside a company to use the same set of standards.

go-micro is a great framework for microservices and very well designed. And it influences micro very much, but there are some coding styles that I don't like go-micro, that's why I made micro for myself.

**Update**: go-micro is now moved to https://github.com/asim/go-micro and it's very well designed. I recommend trying it first to see if it fits your needs.

## Documentation

- See [doc](https://pkg.go.dev/mod/github.com/pthethanh/micro) for package and API descriptions.
- Examples can be found in the [examples](https://github.com/pthethanh/micro/tree/master/examples) directory.
