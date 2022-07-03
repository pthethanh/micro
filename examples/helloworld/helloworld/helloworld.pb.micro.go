// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: helloworld.proto

package helloworld

import (
	fmt "fmt"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (UnimplementedGreeterServer) ServiceDesc() *grpc.ServiceDesc {
	return &Greeter_ServiceDesc
}

func (UnimplementedGreeterServer) RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) {
	RegisterGreeterHandlerFromEndpoint(ctx, mux, endpoint, opts)
}