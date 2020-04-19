package server

import (
	"fmt"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pthethanh/micro/auth"
	"github.com/pthethanh/micro/auth/jwt"
	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultAddr = ":8000"
)

type (
	// Config is a common configuration of a default server.
	// Mostly used by lazy guys via FromEnv().
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
		// MetricsPath is API path for Prometheus metrics.
		MetricsPath string `envconfig:"METRICS_PATH" default:"/internal/metrics"`

		// ReadTimeout is read timeout of both gRPC and HTTP server.
		ReadTimeout time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
		// WriteTimeout is write timeout of both gRPC and HTTP server.
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
		// APIPrefix is path prefix that gRPC API Gateway is routed to.
		APIPrefix string `envconfig:"API_PREFIX" default:"/"`
		// Web is a short config for serving web application.
		// The config format is: path-to-public-dir,index-file-name
		// Example: public,index.html
		Web []string `envconfig:"WEB"`

		// JWTSecret is a short way to enable JWT Authentictor with the secret.
		JWTSecret string `envconfig:"JWT_SECRET"`
		// ContextLogger is an option to enable context logger with request-id.
		ContextLogger bool `envconfig:"CONTEXT_LOGGER" default:"true"`
	}
)

// FromEnv is an option allows user to load configuration from environment variables.
// See Config for the available options.
func FromEnv(configOpts ...config.ReadOption) Option {
	conf := Config{}
	envconfig.Read(&conf, configOpts...)
	return func(server *Server) {
		opts := []Option{
			// Override ADDRESS if PORT is set, mostly for cloud.
			AddressFromEnv(),
			MetricsPaths(conf.ReadinessPath, conf.LivenessPath, conf.MetricsPath),
			TLS(conf.TLSKeyFile, conf.TLSCertFile),
			Timeout(conf.ReadTimeout, conf.WriteTimeout),
			AuthJWT(conf.JWTSecret),
			APIPrefix(conf.APIPrefix),
		}
		if len(conf.Web) == 2 {
			opts = append(opts, Web(conf.Web[0], conf.Web[1]))
		}
		if conf.ContextLogger {
			logger := log.Root()
			if conf.Name != "" {
				logger = logger.Fields("name", conf.Name)
			}
			opts = append(opts, Logger(logger))
		}
		for _, opt := range opts {
			opt(server)
		}
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

// AuthJWT is an option allows user to add jwt authenticator to the server.
func AuthJWT(secret string) Option {
	return func(opts *Server) {
		if secret == "" {
			return
		}
		f := jwt.Authenticator([]byte(secret))
		opts.streamInterceptors = append(opts.streamInterceptors, auth.StreamInterceptor(f))
		opts.unaryInterceptors = append(opts.unaryInterceptors, auth.UnaryInterceptor(f))
	}
}

// Auth is an option allows user to add an authenticator to the server.
func Auth(f auth.AuthenticatorFunc) Option {
	return func(opts *Server) {
		opts.streamInterceptors = append(opts.streamInterceptors, auth.StreamInterceptor(f))
		opts.unaryInterceptors = append(opts.unaryInterceptors, auth.UnaryInterceptor(f))
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
func TLS(key, cert string) Option {
	return func(opts *Server) {
		if key == "" || cert == "" {
			return
		}
		opts.tlsKeyFile = key
		opts.tlsCertFile = cert
		creds, err := credentials.NewServerTLSFromFile(opts.tlsCertFile, opts.tlsKeyFile)
		if err != nil {
			panic(err)
		}
		opts.serverOptions = append(opts.serverOptions, grpc.Creds(creds))
	}
}

// MetricsPaths is an option allows user to override readiness, liveness and metrics path.
func MetricsPaths(ready, live, metrics string) Option {
	return func(opts *Server) {
		opts.readinessPath = ready
		opts.livenessPath = live
		opts.metricsPath = metrics
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
// different path prefix different from root path / by using server.APIPrefix(prefix),
// otherwise a panic will be thrown.
func Web(dir string, index string) Option {
	return func(opts *Server) {
		HTTPHandler("/", spaHandler{
			index: index,
			dir:   dir,
		})(opts)
	}
}

// DefaultHeaderMatcher is an ServerMuxOption that forward
// header keys request-id, api-key to gRPC Context.
func DefaultHeaderMatcher() runtime.ServeMuxOption {
	return HeaderMatcher([]string{"Request-Id", "Api-Key"})
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
