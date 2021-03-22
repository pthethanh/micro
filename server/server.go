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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

		// HTTP
		readTimeout          time.Duration
		writeTimeout         time.Duration
		shutdownTimeout      time.Duration
		routes               []HandlerOptions
		apiPrefix            string
		httpInterceptors     []func(http.Handler) http.Handler
		routesPrioritization bool

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
		server.healthSrv = health.NewServer(map[string]health.CheckFunc{})
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
// If the context is canceled or times out, the gRPC server will attempt a graceful shutdown.
func (server *Server) ListenAndServeContext(ctx context.Context, services ...Service) error {
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
		server.log.Context(ctx).Warn("server: insecured mode is enabled.")
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	// expose health services via gRPC.
	services = append(services, server.healthSrv)

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
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	handler := grpcHandlerFunc(isSecured, grpcServer, router)
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

	// init health check service.
	if err := server.healthSrv.Init(health.StatusServing); err != nil {
		server.log.Context(ctx).Errorf("server: start health check server, err: %v", err)
		server.gracefulShutdown(httpServer, server.shutdownTimeout)
		return err
	}
	defer server.healthSrv.Close()
	server.log.Context(ctx).Infof("server: listening at: %s", server.address)
	select {
	case <-ctx.Done():
		server.gracefulShutdown(httpServer, server.shutdownTimeout)
		return ctx.Err()
	case err := <-errChan:
		return err
	case s := <-sigChan:
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			server.log.Context(ctx).Info("server: gracefully shutdown...")
			server.gracefulShutdown(httpServer, server.shutdownTimeout)
		case os.Kill, syscall.SIGKILL:
			server.log.Context(ctx).Info("server: kill...")
			// It's a kill request, give the server maximum 5s to shutdown.
			t := 5 * time.Second
			if t > server.shutdownTimeout && server.shutdownTimeout > 0 {
				t = server.shutdownTimeout
			}
			server.gracefulShutdown(httpServer, t)
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

// gracefulShutdown shutdown the server gracefully, but with time limit.
// negative timeout is considered as no timeout.
func (server *Server) gracefulShutdown(srv *http.Server, t time.Duration) {
	// tell clients and other services are shutting down...
	// so that no requests should be routed to us.
	if err := server.healthSrv.Close(); err != nil {
		server.log.Errorf("server: shutdown health check service error: %v", err)
	}
	ctx := context.TODO()
	if t >= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), t)
		defer cancel()
	}
	if err := srv.Shutdown(ctx); err != nil {
		server.log.Errorf("server: shutdown error: %v", err)
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
		sort.Slice(server.routes, func(i, j int) bool {
			v := len(strings.Split(server.routes[i].p, "/")) - len(strings.Split(server.routes[j].p, "/"))
			if v != 0 {
				return v > 0
			}
			return strings.Compare(server.routes[i].p, server.routes[j].p) > 0
		})
	}
	for _, r := range server.routes {
		var route *mux.Route
		h := r.h
		for _, interceptor := range r.interceptors {
			h = interceptor(h)
		}
		if r.prefix {
			route = router.PathPrefix(r.p).Handler(h)
			server.log.Context(ctx).Infof("server: registered HTTP handler, path prefix: %s", r.p)
		} else {
			route = router.Path(r.p).Handler(h)
			server.log.Context(ctx).Infof("server: registered HTTP handler, path: %s", r.p)
		}
		if r.m != nil {
			route.Methods(r.m...)
		}
		if r.q != nil {
			route.Queries(r.q...)
		}
		if r.hdr != nil {
			route.Headers(r.hdr...)
		}
	}
}

func isGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}
