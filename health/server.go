package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/pthethanh/micro/log"
)

type (
	// MServer is a simple implementation of Server.
	MServer struct {
		checks map[string]Checker
		ticker *time.Ticker
		log    log.Logger

		server *health.Server
		conf   Config
	}

	// Config hold server config.
	Config struct {
		Interval time.Duration `envconfig:"HEALTH_CHECK_INTERVAL" default:"60s"`
		Timeout  time.Duration `envconfig:"HEALTH_CHECK_TIMEOUT" default:"1s"`
	}

	// ServerOption is a function to provide additional options for server.
	ServerOption func(srv *MServer)

	// Server provides health check services via both gRPC and HTTP.
	// The implementation must follow protocol defined in https://github.com/grpc/grpc/blob/master/doc/health-checking.md
	Server interface {
		// Implements grpc_health_v1.HealthServer for general health check and
		// load balancing according to gRPC protocol.
		grpc_health_v1.HealthServer
		// Implements http.Handler Check API via HTTP.
		http.Handler
		// Register register the Server with grpc.Server
		Register(srv *grpc.Server)
		// Init initialize status, perform necessary setup and start a
		// first health check immediately to update overall status and all
		// dependent services's status.
		Init(status Status) error
		// Close close the underlying resources.
		// It sets all serving status to NOT_SERVING, and configures the server to
		// ignore all future status changes.
		Close() error
	}

	// StatusSetter is an interface to set status for a service according to gRPC Health Check protocol.
	StatusSetter interface {
		// SetStatus is called when need to reset the serving status of a service
		// or insert a new service entry into the statusMap.
		// Use empty string for setting overall status.
		SetStatus(name string, status Status)
	}

	// Status is an alias of grpc_health_v1.HealthCheckResponse_ServingStatus.
	Status = grpc_health_v1.HealthCheckResponse_ServingStatus
	// CheckRequest is an alias of grpc_health_v1.HealthCheckRequest.
	CheckRequest = grpc_health_v1.HealthCheckRequest
	// CheckResponse is an alias of grpc_health_v1.HealthCheckResponse
	CheckResponse = grpc_health_v1.HealthCheckResponse
	// WatchServer is an alias of grpc_health_v1.Health_WatchServer.
	WatchServer = grpc_health_v1.Health_WatchServer
)

// Status const defines short name for grpc_health_v1.HealthCheckResponse_ServingStatus.
const (
	StatusUnknown        = grpc_health_v1.HealthCheckResponse_UNKNOWN
	StatusServing        = grpc_health_v1.HealthCheckResponse_SERVING
	StatusNotServing     = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	StatusServiceUnknown = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
)

const (
	// OverallServiceName is service name of server's overall status.
	OverallServiceName = ""
)

var (
	// force MServer implements required interfaces.
	_ Server       = &MServer{}
	_ StatusSetter = &MServer{}
)

// NewServer return new gRPC health server.
func NewServer(m map[string]Checker, opts ...ServerOption) *MServer {
	srv := &MServer{
		checks: m,
		server: health.NewServer(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	// default if not set
	if srv.conf.Interval == 0 {
		srv.conf.Interval = 60 * time.Second
	}
	if srv.log == nil {
		srv.log = log.Root()
	}
	if srv.conf.Timeout == 0 {
		srv.conf.Timeout = 1 * time.Second
	}
	srv.ticker = time.NewTicker(srv.conf.Interval)

	return srv
}

// Register implements health.Server.
func (s *MServer) Register(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, s)
}

// Init implements health.Server.
func (s *MServer) Init(status Status) error {
	s.server.SetServingStatus(OverallServiceName, status)
	// if there is no dependent services, don't need to do anything else.
	if len(s.checks) == 0 {
		return nil
	}
	// if there are dependent services, set overall status and all dependent services
	// to NotServing as we don't know their status yet.
	s.server.SetServingStatus(OverallServiceName, StatusNotServing)
	for name := range s.checks {
		s.server.SetServingStatus(name, StatusNotServing)
	}
	// start a first check immediately.
	s.checkAll()
	// schedule the check
	go func() {
		for range s.ticker.C {
			s.checkAll()
		}
	}()
	return nil
}

func (s *MServer) checkAll() {
	logger := s.log.Fields(log.CorrelationID, uuid.New().String())
	bg := time.Now()
	overall := StatusServing
	for service, check := range s.checks {
		status := StatusServing
		if err := s.check(service, check); err != nil {
			overall = StatusNotServing
			status = StatusNotServing
			logger.Infof("health check failed, service: %s, err: %v", service, err)
		}
		s.server.SetServingStatus(service, status)
	}
	s.server.SetServingStatus(OverallServiceName, overall)
	logger.Fields("status", overall, "duration", time.Since(bg)).Info("health check completed")
}

func (s *MServer) check(service string, check Checker) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Timeout)
	defer cancel()
	ch := make(chan error)
	go func() {
		ch <- check.CheckHealth(ctx)
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return errors.New("health: check exceed timeout")
	}
}

// Check implements health.Server.
func (s *MServer) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	return s.server.Check(ctx, req)
}

// Watch implements health.Server.
func (s *MServer) Watch(req *CheckRequest, srv WatchServer) error {
	return s.server.Watch(req, srv)
}

// ServeHTTP implements health.Server.
func (s *MServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	data := make(map[string]interface{})
	check := func(service string) Status {
		rs, err := s.Check(r.Context(), &CheckRequest{
			Service: service,
		})
		if status.Code(err) == codes.NotFound {
			s.log.Errorf("health check failed, service: %s, err: %v", service, err)
			return StatusServiceUnknown
		}
		if err != nil {
			s.log.Errorf("health check failed, service: %s, err: %v", service, err)
			return StatusNotServing
		}
		return rs.Status
	}
	// overall - check all dependent services.
	if service == OverallServiceName {
		overall := check(OverallServiceName)
		services := make(map[string]Status)
		for service := range s.checks {
			status := check(service)
			services[service] = status
			if status != StatusServing {
				overall = StatusNotServing
			}
		}
		data["status"] = overall
		data["services"] = services

	} else {
		data["status"] = check(service)
	}
	b, err := json.Marshal(data)
	if err != nil {
		b = []byte(fmt.Sprintf(`{"status":%d}`, StatusNotServing))
		s.log.Errorf("failed to marshal data, err: %v", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// SetStatus implements health.Server
func (s *MServer) SetStatus(service string, status Status) {
	s.server.SetServingStatus(service, status)
}

// Close implements health.Server.
func (s *MServer) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.server != nil {
		s.server.Shutdown()
	}
	return nil
}
