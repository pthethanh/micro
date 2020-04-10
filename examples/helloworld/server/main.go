package main

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	pb "github.com/pthethanh/micro/examples/helloworld/helloworld"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	service struct {
	}
)

// SayHello implements pb.GreeterServer interface.
func (s *service) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name must not be empty")
	}
	return &pb.HelloReply{
		Message: "Hello " + req.GetName(),
	}, nil
}

// Register implements server.Service interface.
func (s *service) Register(srv *grpc.Server) {
	pb.RegisterGreeterServer(srv, s)
}

// RegisterWithEndpoint implements server.EndpointService interface.
func (s *service) RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) {
	pb.RegisterGreeterHandlerFromEndpoint(ctx, mux, addr, opts)
}

func main() {
	srv := &service{}
	if err := server.ListenAndServe(srv); err != nil {
		log.Panic(err)
	}
}
