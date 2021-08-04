// Package client define some utilities for dialing to target gRPC server.
package client

import (
	"context"
	"fmt"
	"os"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/pthethanh/micro/auth/jwt"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/util/contextutil"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/health" // enable health check
	"google.golang.org/grpc/metadata"
)

type (
	// Config hold some basic client configuration.
	Config struct {
		Address     string `envconfig:"ADDRESS" default:"localhost:8000"`
		TLSCertFile string `envconfig:"TLS_CERT_FILE"`
		JWTToken    string `envconfig:"JWT_TOKEN"`
	}

	addressOption struct {
		grpc.EmptyDialOption
		Address string
	}
)

// ReadConfigFromEnv read client config from environment variables.
func ReadConfigFromEnv(opts ...config.ReadOption) *Config {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return &conf
}

// DialOptionsFromEnv return dial options from environment variables.
func DialOptionsFromEnv(opts ...config.ReadOption) []grpc.DialOption {
	return DialOptionsFromConfig(ReadConfigFromEnv(opts...))
}

// DialOptionsFromConfig return dial options from the given configuration.
func DialOptionsFromConfig(conf *Config) []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)
	if conf.JWTToken != "" {
		opts = append(opts, WithJWTCredentials(conf.JWTToken))
	}
	if conf.TLSCertFile != "" {
		opts = append(opts, WithTLSTransportCredentials(conf.TLSCertFile))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if conf.Address != "" {
		opts = append(opts, addressOption{Address: conf.Address})
	}
	return opts
}

// Dial creates a client connection to the given target with health check enabled
// and some others default configurations.
func Dial(address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), address, options...)
}

// DialContext creates a client connection to the given target with health check enabled
// and some others default configurations.
func DialContext(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	for _, opt := range options {
		if addr, ok := opt.(addressOption); ok {
			address = addr.Address
			break
		}
	}
	if address == "" {
		address = GetAddressFromEnv()
	}
	opts := append([]grpc.DialOption{}, options...)
	if len(opts) == 0 {
		opts = append(opts, grpc.WithInsecure())
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
	if p := os.Getenv("PORT"); p != "" {
		return fmt.Sprintf(":%s", p)
	}
	if addr := os.Getenv("ADDRESS"); addr != "" {
		return addr
	}
	return "localhost:8000"
}

// NewContext return new out going context with the given metadata,
// it also copies all associated metadata in the incoming/outcomming context to the new context.
// NewContext panics if len(kv) is odd.
func NewContext(ctx context.Context, kv ...string) context.Context {
	md := metadata.Pairs(kv...)
	if imd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(md, imd)
	}
	if omd, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(md, omd)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// NewTracingContext return new context with the given correlation id for log tracing,
// it also copies all associated metadata in the incoming/outcoming context to the new context.
// If the given correlationID is empty, a new correlation id will be generated.
//
// NOTE: that this function has nothing to do with tracing using opentracing.
func NewTracingContext(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return NewContext(ctx)
	}
	return NewContext(ctx, contextutil.XCorrelationID, correlationID)
}
