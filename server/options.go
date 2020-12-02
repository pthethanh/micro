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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
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

		// LivenessPath is API path for the liveness/health check API.
		LivenessPath string `envconfig:"LIVENESS_PATH" default:"/internal/liveness"`
		// ReadinessPath is API path for the readiness API.
		ReadinessPath string `envconfig:"READINESS_PATH" default:"/internal/readiness"`

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
		Recovery bool `encvonfig:"RECOVERY" default:"true"`

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
			HealthCheckPaths(conf.ReadinessPath, conf.LivenessPath),
			TLS(conf.TLSKeyFile, conf.TLSCertFile),
			Timeout(conf.ReadTimeout, conf.WriteTimeout),
			JWT(conf.JWTSecret),
			APIPrefix(conf.APIPrefix),
			CORS(conf.CORSAllowedCredential, conf.CORSAllowedHeaders, conf.CORSAllowedMethods, conf.CORSAllowedOrigins),
			ShutdownTimeout(conf.ShutdownTimeout),
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

// HealthCheckPaths is an option allows user to override readiness, liveness paths.
func HealthCheckPaths(ready, live string) Option {
	return func(opts *Server) {
		opts.readinessPath = ready
		opts.livenessPath = live
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

// HealthChecks is an option allows user to set health check function.
func HealthChecks(checks ...health.CheckFunc) Option {
	return func(opts *Server) {
		opts.healthChecks = append(opts.healthChecks, checks...)
	}
}

// AddressFromEnv is an option allows user to set address using environment configuration.
// It looks for PORT and then ADDRESS variables.
// This option is mostly used for cloud environment like Heroku where the port
// is randomly set.
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

// HTTPHandler is an option allows user to add additional HTTP handlers.
// If you want to apply middlewares on the HTTP handlers, do it yourselves.
func HTTPHandler(path string, h http.Handler) Option {
	return func(opts *Server) {
		opts.routes = append(opts.routes, route{p: path, h: h})
	}
}

// HTTPInterceptors is an option allows user to set additional interceptors to the root HTTP handler.
// These interceptors are applied to both gRPC and HTTP requests of both public and internal APIs.
func HTTPInterceptors(interceptors ...func(http.Handler) http.Handler) Option {
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
	return func(opts *Server) {
		HTTPHandler(pathPrefix, spaHandler{
			index: index,
			dir:   dir,
		})(opts)
	}
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
		opts.routes = append(opts.routes, route{
			p: pathPrefix + "/debug/pprof/",
			h: http.HandlerFunc(pprof.Index),
		})
		opts.routes = append(opts.routes, route{
			p: pathPrefix + "/debug/pprof/cmdline",
			h: http.HandlerFunc(pprof.Cmdline),
		})
		opts.routes = append(opts.routes, route{
			p: pathPrefix + "/debug/pprof/profile",
			h: http.HandlerFunc(pprof.Profile),
		})
		opts.routes = append(opts.routes, route{
			p: pathPrefix + "/debug/pprof/symbol",
			h: http.HandlerFunc(pprof.Symbol),
		})
		opts.routes = append(opts.routes, route{
			p: pathPrefix + "/debug/pprof/trace",
			h: http.HandlerFunc(pprof.Trace),
		})
	}
}

// Metrics is an option to register standard Prometheus metrics for HTTP.
func Metrics(path string) Option {
	return func(opts *Server) {
		opts.enableMetrics = true
		opts.routes = append(opts.routes, route{
			p: path,
			h: promhttp.Handler(),
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
