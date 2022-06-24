package main

import (
	"context"

	pb "github.com/pthethanh/micro/examples/helloworld/helloworld"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"
	"github.com/pthethanh/micro/status"
)

type (
	service struct {
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

func main() {
	srv := &service{}
	opts := []server.Option{
		server.FromEnv(),
		server.Web("/", "public", "index.html"),
	}
	if err := server.New(opts...).ListenAndServe(srv); err != nil {
		log.Panic(err)
	}
}
