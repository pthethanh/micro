package server

import (
	"context"

	// register default codecs
	_ "github.com/pthethanh/micro/encoding/json"
	_ "google.golang.org/grpc/encoding/proto"
)

// ListenAndServe create a new server base on environment configuration (see server.Config)
// and serve the services with background context.
// See server.ListenAndServe for detail document.
func ListenAndServe(services ...Service) error {
	return ListenAndServeContext(context.Background(), services...)
}

// ListenAndServeContext create a new server base on environment configuration (see server.Config)
// and serve the services with the given context.
// See server.ListenAndServeContext for detail document.
func ListenAndServeContext(ctx context.Context, services ...Service) error {
	return New(FromEnv()).ListenAndServeContext(ctx, services...)
}
