package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type (
	// Service implements a registration interface for services to attach
	// themselves to the grpc.Server.
	Service interface {
		Register(srv *grpc.Server)
	}

	// EndpointService implement an endpoint registration interface for service to attach their endpoint to GRPC gateway
	EndpointService interface {
		RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption)
	}

	// Authenticator defines the interface to perform the actual
	// authentication of the request. Implementations should fetch
	// the required data from the context.Context object. GRPC specific
	// data like `metadata` and `peer` is available on the context.
	// Should return a new `context.Context` that is a child of `ctx`
	// or `codes.Unauthenticated` when auth is lacking or
	// `codes.PermissionDenied` when lacking permissions.
	Authenticator interface {
		Authenticate(ctx context.Context) (context.Context, error)
	}
)

// ListenAndServe opens a tcp listener used by the grpc.Server, and registers
// each Service with the grpc.Server.
func ListenAndServe(conf *Config, services ...Service) error {
	return ListenAndServeContext(context.Background(), conf, services...)
}

// ListenAndServeContext opens a tcp listener used by a grpc.Server and a HTTP server,
// and registers each Service with the grpc.Server. If the Service implements EndpointService
// its endpoints will be registered to the HTTP Server running on the same port.
// If the context is canceled or times out, the GRPC server will attempt a graceful shutdown.
func ListenAndServeContext(ctx context.Context, conf *Config, services ...Service) error {
	lis, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return err
	}
	opts := conf.ServerOptions
	streamInterceptors := []grpc.StreamServerInterceptor{grpc_prometheus.StreamServerInterceptor}
	unaryInterceptors := []grpc.UnaryServerInterceptor{grpc_prometheus.UnaryServerInterceptor}
	isSecured := conf.TLSCertFile != "" && conf.TLSKeyFile != ""

	if conf.Auth != nil {
		streamInterceptors = append(streamInterceptors, auth.StreamInterceptor(conf.Auth))
		unaryInterceptors = append(unaryInterceptors, auth.UnaryInterceptor(conf.Auth))
	}
	if conf.EnableContextLogger {
		logger := log.New(log.Fields{"address": conf.Address})
		streamInterceptors = append(streamInterceptors, log.StreamInterceptor(logger))
		unaryInterceptors = append(unaryInterceptors, log.UnaryInterceptor(logger))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)))
	}

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)))
	}
	if isSecured {
		creds, err := credentials.NewServerTLSFromFile(conf.TLSCertFile, conf.TLSKeyFile)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(opts...)
	muxOpts := conf.ServeMuxOptions
	if len(muxOpts) == 0 {
		muxOpts = []runtime.ServeMuxOption{DefaultHeaderMatcher()}
	}
	gwMux := runtime.NewServeMux(muxOpts...)
	mux := http.NewServeMux()

	dialOpts := make([]grpc.DialOption, 0)
	if isSecured {
		creds, err := credentials.NewClientTLSFromFile(conf.TLSCertFile, "")
		if err != nil {
			return err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	if !isSecured {
		log.WithContext(ctx).Warn("server: insecured mode is enabled.")
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	for i, s := range services {
		s.Register(grpcServer)
		if epSrv, ok := s.(EndpointService); ok {
			log.WithContext(ctx).Infof("server: register HTTP endpoints for service %d", i)
			epSrv.RegisterWithEndpoint(ctx, gwMux, conf.Address, dialOpts)
		}
	}
	// Make sure Prometheus metrics are initialized.
	grpc_prometheus.Register(grpcServer)

	// Attach HTTP handlers
	mux.Handle("/", gwMux)
	mux.Handle(conf.ReadinessPath, health.Readiness())
	mux.Handle(conf.LivenessPath, health.Liveness(conf.HealthChecks...))
	mux.Handle(conf.MetricsPath, promhttp.Handler())

	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	httpServer := &http.Server{
		Addr:         conf.Address,
		Handler:      grpcHandlerFunc(isSecured, grpcServer, mux),
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	go func() {
		if isSecured {
			errChan <- httpServer.ServeTLS(lis, conf.TLSCertFile, conf.TLSKeyFile)
			return
		}
		errChan <- httpServer.Serve(lis)
	}()

	// tell everyone we're ready
	health.Ready()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-errChan:
		return err
	case s := <-sigChan:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			log.WithContext(ctx).Info("server: gracefully shutdown...")
			grpcServer.GracefulStop()
		case os.Kill, syscall.SIGKILL:
			log.WithContext(ctx).Info("server: kill...")
			grpcServer.Stop()
		}
		// waiting for srv.Serve to return to errChan.
	}
	return nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func grpcHandlerFunc(isSecured bool, grpcServer *grpc.Server, mux http.Handler) http.Handler {
	if isSecured {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
				return
			}
			mux.ServeHTTP(w, r)
		})
	}
	// work-around in case TLS is disabled. See: https://github.com/grpc/grpc-go/issues/555
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
