package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"net/textproto"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
)

const (
	defaultAddr = ":8000"
)

type (
	// Config is a common configuration of a default server.
	Config struct {
		// Name is name of the service.
		Name string `envconfig:"NAME" default:"micro"`
		// Address is the address of the service in form of host:port.
		Address string `envconfig:"ADDRESS" default:":8000"`
		// TLSCertFile is path to the TLS certificate file.
		TLSCertFile string `envconfig:"TLS_CERT_FILE"`
		// TLSKeyFile is the path to the TLS key file.
		TLSKeyFile string `envconfig:"TLS_KEY_FILE"`

		// HealthCheckPath is API path for the health check.
		HealthCheckPath string `envconfig:"HEALTH_CHECK_PATH" default:"/internal/health"`

		// ReadTimeout is read timeout of both gRPC and HTTP server.
		ReadTimeout time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
		// WriteTimeout is write timeout of both gRPC and HTTP server.
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
		//ShutdownTimeout is timeout for shutting down the server.
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
		// APIPrefix is path prefix that gRPC API Gateway is routed to.
		APIPrefix string `envconfig:"API_PREFIX" default:"/api/"`

		// Web options
		WebDir    string `envconfig:"WEB_DIR"`
		WebIndex  string `envconfig:"WEB_INDEX" default:"index.html"`
		WebPrefix string `envconfig:"WEB_PREFIX" default:"/"`

		// JWTSecret is a short way to enable JWT Authentictor with the secret.
		JWTSecret string `envconfig:"JWT_SECRET"`
		// ContextLogger is an option to enable context logger with request-id.
		ContextLogger bool `envconfig:"CONTEXT_LOGGER" default:"true"`

		// Recovery is a short way to enable recovery interceptors for both unary and stream handlers.
		Recovery bool `envconfig:"RECOVERY" default:"true"`

		// CORS options
		CORSAllowedHeaders    []string `envconfig:"CORS_ALLOWED_HEADERS"`
		CORSAllowedMethods    []string `envconfig:"CORS_ALLOWED_METHODS"`
		CORSAllowedOrigins    []string `envconfig:"CORS_ALLOWED_ORIGINS"`
		CORSAllowedCredential bool     `envconfig:"CORS_ALLOWED_CREDENTIAL" default:"false"`

		// PProf options
		PProf       bool   `envconfig:"PPROF" default:"false"`
		PProfPrefix string `envconfig:"PPROF_PREFIX"`

		// Metrics enable/disable standard metrics
		Metrics bool `envconfig:"METRICS" default:"true"`
		// MetricsPath is API path for Prometheus metrics.
		MetricsPath string `envconfig:"METRICS_PATH" default:"/internal/metrics"`

		RoutesPrioritization bool `envconfig:"ROUTES_PRIORITIZATION" default:"true"`
	}
)

// ReadConfigFromEnv read the server configuration from environment variables.
func ReadConfigFromEnv(opts ...config.ReadOption) Config {
	conf := Config{}
	envconfig.Read(&conf, opts...)
	return conf
}

// FromEnv is an option to create a new server from environment variables configuration.
// See Config for the available options.
func FromEnv(configOpts ...config.ReadOption) Option {
	conf := Config{}
	envconfig.Read(&conf, configOpts...)
	return func(opts *Server) {
		FromConfig(conf)(opts)
		AddressFromEnv()(opts)
	}
}

// FromConfig is an option to create a new server from an existing config.
func FromConfig(conf Config) Option {
	return func(server *Server) {
		opts := []Option{
			Address(conf.Address),
			TLS(conf.TLSKeyFile, conf.TLSCertFile),
			Timeout(conf.ReadTimeout, conf.WriteTimeout),
			JWT(conf.JWTSecret),
			APIPrefix(conf.APIPrefix),
			CORS(conf.CORSAllowedCredential, conf.CORSAllowedHeaders, conf.CORSAllowedMethods, conf.CORSAllowedOrigins),
			ShutdownTimeout(conf.ShutdownTimeout),
			RoutesPrioritization(conf.RoutesPrioritization),
		}
		if conf.Metrics {
			opts = append(opts, Metrics(conf.MetricsPath))
		}
		if conf.WebDir != "" {
			opts = append(opts, Web(conf.WebPrefix, conf.WebDir, conf.WebIndex))
		}
		// context logger
		if conf.ContextLogger {
			logger := log.Root()
			if conf.Name != "" {
				logger = logger.Fields("name", conf.Name)
			}
			opts = append(opts, Logger(logger))
		}
		// create health check by default
		opts = append(opts, HealthCheck(conf.HealthCheckPath,
			health.NewServer(map[string]health.CheckFunc{},
				health.Logger(server.getLogger()))))
		// recovery
		if conf.Recovery {
			opts = append(opts, Recovery(nil))
		}
		if conf.PProf {
			opts = append(opts, PProf(conf.PProfPrefix))
		}
		// apply all
		for _, opt := range opts {
			opt(server)
		}
	}
}

// Address is an option to set address.
// Default address is :8000
func Address(addr string) Option {
	return func(opts *Server) {
		opts.address = addr
		if opts.lis != nil {
			opts.getLogger().Debugf("server: address is set to %s, Listener will be overridden", addr)
			opts.lis = nil
		}
	}
}

// Listener is an option allows server to be served on an existing listener.
func Listener(lis net.Listener) Option {
	return func(opts *Server) {
		if opts.address != "" {
			opts.getLogger().Debugf("server: listener is set to %s, address will be overridden", lis.Addr().String())
			opts.address = lis.Addr().String()
		}
		opts.lis = lis
	}
}

// StreamInterceptors is an option allows user to add additional stream interceptors to the server.
func StreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(opts *Server) {
		opts.streamInterceptors = append(opts.streamInterceptors, interceptors...)
	}
}

// UnaryInterceptors is an option allows user to add additional unary interceptors to the server.
func UnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(opts *Server) {
		opts.unaryInterceptors = append(opts.unaryInterceptors, interceptors...)
	}
}

// JWT is an option allows user to use jwt authenticator for authentication.
func JWT(secret string) Option {
	return func(opts *Server) {
		if secret == "" {
			return
		}
		opts.auth = jwt.Authenticator([]byte(secret))
	}
}

// Auth is an option allows user to use an authenticator for authentication.
// Find more about authenticators in auth package.
func Auth(f auth.Authenticator) Option {
	return func(opts *Server) {
		opts.auth = f
	}
}

// Logger is an option allows user to add a custom logger into the server.
func Logger(logger log.Logger) Option {
	return func(opts *Server) {
		opts.log = logger
		opts.serveMuxOptions = append(opts.serveMuxOptions, DefaultHeaderMatcher())
		opts.streamInterceptors = append(opts.streamInterceptors, log.StreamInterceptor(logger))
		opts.unaryInterceptors = append(opts.unaryInterceptors, log.UnaryInterceptor(logger))
	}
}

// TLS is an option allows user to add TLS for transport security to the server.
// Note that host name in ADDRESS must be configured accordingly. Otherwise, you
// might encounter TLS handshake error.
// TIP: for local testing, take a look at https://github.com/FiloSottile/mkcert.
func TLS(key, cert string) Option {
	return func(opts *Server) {
		if key == "" || cert == "" {
			return
		}
		opts.tlsKeyFile = key
		opts.tlsCertFile = cert
		// server/dial options will be handled in server.go
	}
}

// Timeout is an option to override default read/write timeout.
func Timeout(read, write time.Duration) Option {
	if read == 0 {
		read = 30 * time.Second
	}
	if write == 0 {
		write = 30 * time.Second
	}
	return func(opts *Server) {
		opts.readTimeout = read
		opts.writeTimeout = write
		opts.serverOptions = append(opts.serverOptions, grpc.ConnectionTimeout(read))
	}
}

// ServeMuxOptions is an option allows user to add additional ServeMuxOption.
func ServeMuxOptions(muxOpts ...runtime.ServeMuxOption) Option {
	return func(opts *Server) {
		opts.serveMuxOptions = append(opts.serveMuxOptions, muxOpts...)
	}
}

// Options is an option allows user to add additional grpc.ServerOption.
func Options(serverOpts ...grpc.ServerOption) Option {
	return func(opts *Server) {
		opts.serverOptions = append(opts.serverOptions, serverOpts...)
	}
}

// HealthCheck is an option allows user to provide custom health check server.
func HealthCheck(path string, srv health.Server) Option {
	return func(opts *Server) {
		opts.healthCheckPath = path
		opts.healthSrv = srv
	}
}

// AddressFromEnv is an option allows user to set address using environment configuration.
// It looks for PORT and then ADDRESS variables.
// This option is mostly used for cloud environment like Heroku where the port is randomly set.
func AddressFromEnv() Option {
	return func(opts *Server) {
		if p := os.Getenv("PORT"); p != "" {
			opts.address = fmt.Sprintf(":%s", p)
			return
		}
		if addr := os.Getenv("ADDRESS"); addr != "" {
			opts.address = addr
			return
		}
		if opts.address == "" {
			opts.address = defaultAddr
		}
	}
}

// Handler is an option allows user to add additional HTTP handlers.
// Longer patterns take precedence over shorter ones by default,
// use RoutesPrioritization option to disable this rule.
// See github.com/gorilla/mux for defining path with variables/patterns.
//
// For more options, use HTTPHandlerX.
func Handler(path string, h http.Handler, methods ...string) Option {
	return HandlerWithOptions(path, h, NewHandlerOptions().Methods(methods...))
}

// HandlerFunc is an option similar to HTTPHandler, but for http.HandlerFunc.
func HandlerFunc(path string, h func(http.ResponseWriter, *http.Request), methods ...string) Option {
	return HandlerWithOptions(path, http.HandlerFunc(h), NewHandlerOptions().Methods(methods...))
}

// PrefixHandler is an option to quickly define a prefix HTTP handler.
// For more options, please use HTTPHandlerX.
func PrefixHandler(path string, h http.Handler, methods ...string) Option {
	return HandlerWithOptions(path, h, NewHandlerOptions().Prefix().Methods(methods...))
}

// HandlerWithOptions is an option to define full options such as method, query, header matchers
// and interceptors for a HTTP handler.
// Longer patterns take precedence over shorter ones by default,
// use RoutesPrioritization option to disable this rule.
// See github.com/gorilla/mux for defining path with variables/patterns.
func HandlerWithOptions(path string, h http.Handler, hopt *HandlerOptions) Option {
	return func(opts *Server) {
		hopt.p = path
		hopt.h = h
		opts.routes = append(opts.routes, *hopt)
	}
}

// RoutesPrioritization enable/disable the routes prioritization.
func RoutesPrioritization(enable bool) Option {
	return func(opts *Server) {
		opts.routesPrioritization = enable
	}
}

// HTTPInterceptors is an option allows user to set additional interceptors to the root HTTP handler.
// If interceptors are applied to gRPC, it is required that the interceptors don't hijack the response writer,
// otherwise panic "Hijack not supported" will be thrown.
func HTTPInterceptors(interceptors ...HTTPInterceptor) Option {
	return func(opts *Server) {
		opts.httpInterceptors = append(opts.httpInterceptors, interceptors...)
	}
}

// APIPrefix is an option allows user to route only the specified path prefix to gRPC Gateway.
// This option is used mostly when you serve both gRPC APIs along with other internal HTTP APIs.
// The default prefix is /, which will route all paths to gRPC Gateway.
func APIPrefix(prefix string) Option {
	return func(opts *Server) {
		opts.apiPrefix = prefix
	}
}

// Web is an option to allows user to serve Web/Single Page Application
// along with API Gateway and gRPC. API Gateway must be served in a
// different path prefix with the web path prefix.
func Web(pathPrefix, dir, index string) Option {
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	return PrefixHandler(pathPrefix, spaHandler{
		index: index,
		dir:   dir,
	})
}

// Recovery is an option allows user to add an ability to recover a handler/API from a panic.
// This applies for both unary and stream handlers/APIs.
// If the provided error handler is nil, a default error handler will be used.
func Recovery(handler func(context.Context, interface{}) error) Option {
	return func(opts *Server) {
		if handler == nil {
			if opts.log == nil {
				opts.log = log.Root()
			}
			handler = recoveryHandler(opts.log)
		}
		recoverOpt := recovery.WithRecoveryHandlerContext(handler)
		opts.unaryInterceptors = append(opts.unaryInterceptors, recovery.UnaryServerInterceptor(recoverOpt))
		opts.streamInterceptors = append(opts.streamInterceptors, recovery.StreamServerInterceptor(recoverOpt))
	}
}

// CORS is an option allows users to enable CORS on the server.
func CORS(allowCredential bool, headers, methods, origins []string) Option {
	return func(opts *Server) {
		options := []handlers.CORSOption{}
		if allowCredential {
			options = append(options, handlers.AllowCredentials())
		}
		if headers != nil {
			options = append(options, handlers.AllowedHeaders(headers))
		}
		if methods != nil {
			options = append(options, handlers.AllowedMethods(methods))
		}
		if origins != nil {
			options = append(options, handlers.AllowedOrigins(origins))
		}
		if len(options) > 0 {
			opts.httpInterceptors = append(opts.httpInterceptors, handlers.CORS(options...))
		}
	}
}

// PProf is an option allows user to enable Go profiler.
func PProf(pathPrefix string) Option {
	return func(opts *Server) {
		opts.routes = append(opts.routes, HandlerOptions{
			p: pathPrefix + "/debug/pprof/",
			h: http.HandlerFunc(pprof.Index),
		})
		opts.routes = append(opts.routes, HandlerOptions{
			p: pathPrefix + "/debug/pprof/cmdline",
			h: http.HandlerFunc(pprof.Cmdline),
		})
		opts.routes = append(opts.routes, HandlerOptions{
			p: pathPrefix + "/debug/pprof/profile",
			h: http.HandlerFunc(pprof.Profile),
		})
		opts.routes = append(opts.routes, HandlerOptions{
			p: pathPrefix + "/debug/pprof/symbol",
			h: http.HandlerFunc(pprof.Symbol),
		})
		opts.routes = append(opts.routes, HandlerOptions{
			p: pathPrefix + "/debug/pprof/trace",
			h: http.HandlerFunc(pprof.Trace),
		})
	}
}

// Metrics is an option to register standard Prometheus metrics for HTTP.
// Default path is /internal/metrics.
func Metrics(path string) Option {
	return func(opts *Server) {
		p := path
		if p == "" {
			p = "/internal/metrics"
			opts.getLogger().Warnf("metrics path is switched automatically to %s", p)
		}
		opts.enableMetrics = true
		opts.routes = append(opts.routes, HandlerOptions{
			p: p,
			h: promhttp.Handler(),
			m: []string{http.MethodGet},
		})
	}
}

// ShutdownTimeout is an option to override default shutdown timeout of server.
// Set to -1 for no timeout.
func ShutdownTimeout(t time.Duration) Option {
	return func(opts *Server) {
		opts.shutdownTimeout = t
	}
}

// DefaultHeaderMatcher is an ServerMuxOption that forward
// header keys X-Request-Id, X-Correlation-ID, Api-Key to gRPC Context.
func DefaultHeaderMatcher() runtime.ServeMuxOption {
	return HeaderMatcher([]string{"X-Request-Id", "X-Correlation-ID", "Api-Key"})
}

// HeaderMatcher is an ServeMuxOption for matcher header
// for passing a set of non IANA headers to gRPC context
// without a need to prefix them with Grpc-Metadata.
func HeaderMatcher(keys []string) runtime.ServeMuxOption {
	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, k := range keys {
			if k == canonicalKey || textproto.CanonicalMIMEHeaderKey(k) == canonicalKey {
				return k, true
			}
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}

// recoveryHandler print the context log to the configured writer and return
// a general error to the caller.
func recoveryHandler(l log.Logger) func(context.Context, interface{}) error {
	return func(ctx context.Context, p interface{}) error {
		l.Context(ctx).Errorf("server: panic recovered, err: %v", p)
		return status.Errorf(codes.Internal, codes.Internal.String())
	}
}
