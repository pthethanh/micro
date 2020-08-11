// Package server provides a convenient way to start a new ready to use server with default
// HTTP API for readiness, liveness and Prometheus metrics.
package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type (
	// Server holds the configuration options for the server instance.
	Server struct {
		lis         net.Listener
		address     string
		tlsCertFile string
		tlsKeyFile  string

		// Paths
		livenessPath  string
		readinessPath string

		// HTTP
		readTimeout      time.Duration
		writeTimeout     time.Duration
		routes           []route
		apiPrefix        string
		httpInterceptors []func(http.Handler) http.Handler

		// Needs to be set manually
		healthChecks    []health.CheckFunc
		serverOptions   []grpc.ServerOption
		serveMuxOptions []runtime.ServeMuxOption

		// Interceptors
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor

		log           log.Logger
		enableMetrics bool
	}

	// Option is a configuration option.
	Option func(*Server)

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

	route struct {
		p     string
		h     http.Handler
		proto []string
	}
)

// New return new server with the given options.
// If address is not set, default address ":8000" will be used.
func New(ops ...Option) *Server {
	server := &Server{}
	for _, op := range ops {
		op(server)
	}
	if server.log == nil {
		server.log = log.Root()
	}
	if server.address == "" {
		server.address = defaultAddr
	}
	return server
}

// ListenAndServe call ListenAndServeContext with background context.
func (server *Server) ListenAndServe(services ...Service) error {
	return server.ListenAndServeContext(context.Background(), services...)
}

// ListenAndServeContext opens a tcp listener used by a grpc.Server and a HTTP server,
// and registers each Service with the grpc.Server. If the Service implements EndpointService
// its endpoints will be registered to the HTTP Server running on the same port.
// The server starts with default metrics and health endpoints.
// If the context is canceled or times out, the GRPC server will attempt a graceful shutdown.
func (server *Server) ListenAndServeContext(ctx context.Context, services ...Service) error {
	if server.lis == nil {
		lis, err := net.Listen("tcp", server.address)
		if err != nil {
			return err
		}
		server.lis = lis
	}
	server.streamInterceptors = append(server.streamInterceptors, grpc_prometheus.StreamServerInterceptor)
	server.unaryInterceptors = append(server.unaryInterceptors, grpc_prometheus.UnaryServerInterceptor)
	isSecured := server.tlsCertFile != "" && server.tlsKeyFile != ""

	if len(server.streamInterceptors) > 0 {
		server.serverOptions = append(server.serverOptions, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(server.streamInterceptors...)))
	}
	if len(server.unaryInterceptors) > 0 {
		server.serverOptions = append(server.serverOptions, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(server.unaryInterceptors...)))
	}
	if isSecured {
		creds, err := credentials.NewServerTLSFromFile(server.tlsCertFile, server.tlsKeyFile)
		if err != nil {
			return err
		}
		server.serverOptions = append(server.serverOptions, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(server.serverOptions...)
	muxOpts := server.serveMuxOptions
	if len(muxOpts) == 0 {
		muxOpts = []runtime.ServeMuxOption{DefaultHeaderMatcher()}
	}
	gw := runtime.NewServeMux(muxOpts...)
	mux := http.NewServeMux()

	dialOpts := make([]grpc.DialOption, 0)
	if isSecured {
		creds, err := credentials.NewClientTLSFromFile(server.tlsCertFile, "")
		if err != nil {
			return err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	if !isSecured {
		server.log.Context(ctx).Warn("server: insecured mode is enabled.")
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	for _, s := range services {
		s.Register(grpcServer)
		if epSrv, ok := s.(EndpointService); ok {
			epSrv.RegisterWithEndpoint(ctx, gw, server.address, dialOpts)
		}
	}
	// Make sure Prometheus metrics are initialized.
	if server.enableMetrics {
		grpc_prometheus.Register(grpcServer)
	}
	// Attach handlers by order: internal, HTTP handlers, gRPC.
	server.routes = append([]route{
		{
			p: server.getReadinessPath(),
			h: health.Readiness(),
		},
		{
			p: server.getLivenessPath(),
			h: health.Liveness(server.healthChecks...),
		},
	}, server.routes...)
	server.routes = append(server.routes, route{
		p:     server.getAPIPrefix(),
		h:     gw,
		proto: []string{"HTTP", "gRPC"},
	})

	for _, r := range server.routes {
		proto := strings.Join(r.proto, "+")
		if len(proto) == 0 {
			proto = "HTTP"
		}
		log.Context(ctx).Infof("registered handler, path: %s, proto: %s", r.p, proto)
		mux.Handle(r.p, r.h)
	}

	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	handler := grpcHandlerFunc(isSecured, grpcServer, mux)
	for i := len(server.httpInterceptors) - 1; i >= 0; i-- {
		handler = server.httpInterceptors[i](handler)
	}
	httpServer := &http.Server{
		Addr:         server.address,
		Handler:      handler,
		ReadTimeout:  server.readTimeout,
		WriteTimeout: server.writeTimeout,
	}

	go func() {
		if isSecured {
			errChan <- httpServer.ServeTLS(server.lis, server.tlsCertFile, server.tlsKeyFile)
			return
		}
		errChan <- httpServer.Serve(server.lis)
	}()

	// tell everyone we're ready
	health.Ready()
	server.log.Context(ctx).Infof("listening at: %s", server.address)
	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-errChan:
		return err
	case s := <-sigChan:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			log.Context(ctx).Info("server: gracefully shutdown...")
			grpcServer.GracefulStop()
		case os.Kill, syscall.SIGKILL:
			log.Context(ctx).Info("server: kill...")
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

func (server Server) getReadinessPath() string {
	if server.readinessPath == "" {
		return "/internal/readiness"
	}
	return server.readinessPath
}

func (server Server) getLivenessPath() string {
	if server.livenessPath == "" {
		return "/internal/liveness"
	}
	return server.livenessPath
}

// With allows user to add more options to the server after created.
func (server *Server) With(opts ...Option) *Server {
	for _, op := range opts {
		op(server)
	}
	return server
}

// Address return address that the server is listening.
func (server *Server) Address() string {
	return server.address
}

func (server Server) getAPIPrefix() string {
	if server.apiPrefix == "" {
		return "/"
	}
	return server.apiPrefix
}
