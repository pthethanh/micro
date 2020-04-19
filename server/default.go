package server

import (
	"context"
)

// NewFromEnv load configurations from environment and create a new server.
// Additional options can be added to the sever via Server.WithOptions(...).
// See Config for environment names.
func NewFromEnv() *Server {
	return New(defaultAddr, FromEnv())
}

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
	return NewFromEnv().ListenAndServeContext(ctx, services...)
}
