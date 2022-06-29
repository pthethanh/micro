// Package server provides a convenient way to create or start a new server
// that serves both gRPC and HTTP over 1 single port
// with default useful APIs for authentication, health check, metrics,...
package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/status"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	// Server holds the configuration options for the server instance.
	Server struct {
		lis         net.Listener
		httpSrv     *http.Server
		address     string
		tlsCertFile string
		tlsKeyFile  string

		// HTTP
		readTimeout          time.Duration
		writeTimeout         time.Duration
		shutdownTimeout      time.Duration
		routes               []HandlerOptions
		apiPrefix            string
		httpInterceptors     []func(http.Handler) http.Handler
		routesPrioritization bool
		notFoundHandler      http.Handler

		serverOptions   []grpc.ServerOption
		serveMuxOptions []runtime.ServeMuxOption

		// Interceptors
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor

		log           log.Logger
		enableMetrics bool

		auth auth.Authenticator

		// health checks
		healthCheckPath string
		healthSrv       health.Server
	}

	// Option is a configuration option.
	Option func(*Server)

	// Service implements a registration interface for services to attach themselves to the grpc.Server.
	// If the services support gRPC gateway, they must also implement the EndpointService interface.
	Service interface {
		Register(srv *grpc.Server)
	}

	// ServiceDescriptor implements grpc service that expose its service desc.
	ServiceDescriptor interface {
		ServiceDesc() *grpc.ServiceDesc
	}

	// EndpointService implement an endpoint registration interface for service to attach their endpoints to gRPC gateway.
	EndpointService interface {
		RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption)
	}
)

// New return new server with the given options.
// If address is not set, default address ":8000" will be used.
func New(ops ...Option) *Server {
	server := &Server{
		routesPrioritization: true,
	}
	for _, op := range ops {
		op(server)
	}
	if server.log == nil {
		server.log = log.Root()
	}
	if server.address == "" {
		server.address = defaultAddr
	}
	if server.healthSrv == nil {
		server.healthSrv = health.NewServer(map[string]health.Checker{})
	}

	return server
}

// ListenAndServe call ListenAndServeContext with background context.
func (server *Server) ListenAndServe(services ...interface{}) error {
	return server.ListenAndServeContext(context.Background(), services...)
}

// ListenAndServeContext opens a tcp listener used by a grpc.Server and a HTTP server,
// and registers each Service with the grpc.Server. If the Service implements EndpointService
// its endpoints will be registered to the HTTP Server running on the same port.
// The server starts with default metrics and health endpoints.
// If the context is canceled or times out, the gRPC server will attempt a graceful shutdown.
func (server *Server) ListenAndServeContext(ctx context.Context, services ...interface{}) error {
	if server.lis == nil {
		lis, err := net.Listen("tcp", server.address)
		if err != nil {
			return err
		}
		server.lis = lis
	}
	if server.auth != nil {
		server.streamInterceptors = append(server.streamInterceptors, auth.StreamInterceptor(server.auth))
		server.unaryInterceptors = append(server.unaryInterceptors, auth.UnaryInterceptor(server.auth))
	}
	if server.enableMetrics {
		server.streamInterceptors = append(server.streamInterceptors, grpc_prometheus.StreamServerInterceptor)
		server.unaryInterceptors = append(server.unaryInterceptors, grpc_prometheus.UnaryServerInterceptor)
	}
	isSecured := server.tlsCertFile != "" && server.tlsKeyFile != ""

	// server options
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
	router := mux.NewRouter()

	dialOpts := make([]grpc.DialOption, 0)
	if isSecured {
		creds, err := credentials.NewClientTLSFromFile(server.tlsCertFile, "")
		if err != nil {
			return err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	if !isSecured {
		server.log.Context(ctx).Warn("server: insecure mode is enabled.")
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	// expose health services via gRPC.
	services = append(services, server.healthSrv)

	for _, s := range services {
		c := 0
		if srv, ok := s.(Service); ok {
			srv.Register(grpcServer)
			c++
		} else if srv, ok := s.(ServiceDescriptor); ok {
			grpcServer.RegisterService(srv.ServiceDesc(), srv)
			c++
		}
		if srv, ok := s.(EndpointService); ok {
			srv.RegisterWithEndpoint(ctx, gw, server.address, dialOpts)
			c++
		}
		if c == 0 {
			return status.InvalidArgument("invalid service registration: %v, service should implement one of the interface: server.Service, server.ServiceDescriptor or server.EndpointService", s)
		}
	}
	// Make sure Prometheus metrics are initialized.
	if server.enableMetrics {
		grpc_prometheus.Register(grpcServer)
	}
	// Add internal handlers.
	server.routes = append([]HandlerOptions{
		{
			p: server.getHealthCheckPath(),
			h: server.healthSrv,
			m: []string{http.MethodGet},
		},
	}, server.routes...)
	// Serve gRPC and GW only and only if there is at least one service registered.
	if len(services) > 0 {
		server.routes = append(server.routes, HandlerOptions{p: server.getAPIPrefix(), h: gw, prefix: true})
	}
	// register all http handlers to the router.
	server.registerHTTPHandlers(ctx, router)

	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	handler := grpcHandlerFunc(isSecured, grpcServer, router)
	for i := len(server.httpInterceptors) - 1; i >= 0; i-- {
		handler = server.httpInterceptors[i](handler)
	}
	server.httpSrv = &http.Server{
		Addr:         server.address,
		Handler:      handler,
		ReadTimeout:  server.readTimeout,
		WriteTimeout: server.writeTimeout,
	}
	go func() {
		if isSecured {
			errChan <- server.httpSrv.ServeTLS(server.lis, server.tlsCertFile, server.tlsKeyFile)
			return
		}
		errChan <- server.httpSrv.Serve(server.lis)
	}()

	// init health check service.
	if err := server.healthSrv.Init(health.StatusServing); err != nil {
		server.log.Context(ctx).Errorf("server: start health check server, err: %v", err)
		server.Shutdown(ctx)
		return err
	}
	defer server.healthSrv.Close()
	server.log.Context(ctx).Infof("server: listening at: %s", server.address)
	select {
	case <-ctx.Done():
		server.Shutdown(ctx)
		return ctx.Err()
	case err := <-errChan:
		return err
	case s := <-sigChan:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			server.log.Context(ctx).Info("server: gracefully shutdown...")
			server.Shutdown(ctx)
		}
	}
	return nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func grpcHandlerFunc(isSecured bool, grpcServer *grpc.Server, mux http.Handler) http.Handler {
	if isSecured {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isGRPCRequest(r) {
				grpcServer.ServeHTTP(w, r)
				return
			}
			mux.ServeHTTP(w, r)
		})
	}
	// work-around in case TLS is disabled. See: https://github.com/grpc/grpc-go/issues/555
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGRPCRequest(r) {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func (server Server) getHealthCheckPath() string {
	if server.healthCheckPath == "" {
		return "/internal/health"
	}
	return server.healthCheckPath
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

// Shutdown shutdown the server gracefully.
func (server *Server) Shutdown(ctx context.Context) {
	if server.healthSrv != nil {
		if err := server.healthSrv.Close(); err != nil {
			server.log.Errorf("server: shutdown health check service error: %v", err)
		}
	}
	if server.httpSrv != nil {
		ctx, cancel := context.WithTimeout(ctx, server.shutdownTimeout)
		defer cancel()
		if err := server.httpSrv.Shutdown(ctx); err != nil {
			server.log.Errorf("server: shutdown error: %v", err)
		}
	}
}

func (server *Server) getLogger() log.Logger {
	if server.log == nil {
		return log.Root()
	}
	return server.log
}

func (server *Server) registerHTTPHandlers(ctx context.Context, router *mux.Router) {
	// Longer patterns take precedence over shorter ones.
	if server.routesPrioritization {
		sort.Sort(sort.Reverse(handlerOptionsSlice(server.routes)))
	}
	if server.notFoundHandler != nil {
		router.NotFoundHandler = server.notFoundHandler
	}
	for _, r := range server.routes {
		var route *mux.Route
		h := r.h
		info := make([]interface{}, 0)
		for _, interceptor := range r.interceptors {
			h = interceptor(h)
		}
		if r.prefix {
			route = router.PathPrefix(r.p).Handler(h)
			info = append(info, "path_prefix", r.p)
		} else {
			route = router.Path(r.p).Handler(h)
			info = append(info, "path", r.p)
		}
		if r.m != nil {
			route.Methods(r.m...)
			info = append(info, "methods", r.m)
		}
		if r.q != nil {
			route.Queries(r.q...)
			info = append(info, "queries", r.q)
		}
		if r.hdr != nil {
			route.Headers(r.hdr...)
			info = append(info, "headers", r.hdr)
		}
		server.log.Context(ctx).Fields(info...).Infof("server: registered HTTP handler")

	}
}

func isGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}
