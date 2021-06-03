// Package client define some utilities for dialing to target gRPC server.
package client

import (
	"context"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/pthethanh/micro/auth/jwt"
	"github.com/pthethanh/micro/log"
	"google.golang.org/grpc"

	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/health" // enable health check
)

// Dial creates a client connection to the given target with health check enabled
// and some others default configurations.
func Dial(address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), address, options...)
}

// DialContext creates a client connection to the given target with health check enabled
// and some others default configurations.
func DialContext(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	if address == "" {
		address = GetAddressFromEnv()
	}
	opts := append([]grpc.DialOption{}, options...)
	opts = append(opts, grpc.WithInsecure())
	if len(opts) == 0 {
		opts = append(opts, grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: 30 * time.Second,
		}))
	}
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		log.Errorf("dial to %s failed, err: %v", address, err)
		return nil, err
	}
	return conn, nil
}

// Must return the given client connection if err is nil, otherwise panic.
func Must(conn *grpc.ClientConn, err error) *grpc.ClientConn {
	if err != nil {
		panic(err)
	}
	return conn
}

// WithJWTCredentials return a JWT credential dial option.
func WithJWTCredentials(token string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(jwt.NewJWTCredentialsFromToken(token))
}

// WithTLSTransportCredentials return new dial option using TLS.
// panic if the given file is not found.
func WithTLSTransportCredentials(certFile string) grpc.DialOption {
	opt, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		panic(err)
	}
	return grpc.WithTransportCredentials(opt)
}

// WithTracing return unary tracing interceptor dial option.
func WithTracing(tracer opentracing.Tracer) grpc.DialOption {
	return grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer))
}

// WithStreamTracing return stream tracing interceptor dial option.
func WithStreamTracing(tracer opentracing.Tracer) grpc.DialOption {
	return grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer))
}

// GetAddressFromEnv return address from environment variable ADDRESS if defined.
// otherwise return default :8000
func GetAddressFromEnv() string {
	addr := os.Getenv("ADDRESS")
	if addr != "" {
		return addr
	}
	return ":8000"
}
