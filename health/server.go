package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/google/uuid"
	"github.com/pthethanh/micro/log"
)

type (
	// MServer is a simple implementation of Server.
	MServer struct {
		interval time.Duration
		timeout  time.Duration
		checks   map[string]CheckFunc
		ticker   *time.Ticker
		log      log.Logger

		*health.Server
	}

	// ServerOption is a function to provide additional options for server.
	ServerOption func(srv *MServer)

	// Server provides health check services via both gRPC and HTTP.
	// The implementation must follow protocol defined in https://github.com/grpc/grpc/blob/master/doc/health-checking.md
	Server interface {
		grpc_health_v1.HealthServer
		http.Handler
		ServingStatusSetter
		// Register register the Server with grpc.Server
		Register(srv *grpc.Server)
		// Init inits and perform a first health check immediately
		// to update overall status and all dependent services's status.
		Init(status Status) error
		// Close close the underlying resources.
		// It sets all serving status to NOT_SERVING, and configures the server to
		// ignore all future status changes.
		Close() error
	}

	// ServingStatusSetter is an interface to set status for a service according to gRPC Health Check protocol.
	ServingStatusSetter interface {
		// SetServingStatus is called when need to reset the serving status of a service
		// or insert a new service entry into the statusMap.
		SetServingStatus(name string, status Status)
	}

	// Status is an alias of grpc_health_v1.HealthCheckResponse_ServingStatus.
	Status = grpc_health_v1.HealthCheckResponse_ServingStatus
)

// Status const defines short name for grpc_health_v1.HealthCheckResponse_ServingStatus.
const (
	StatusUnknown        = grpc_health_v1.HealthCheckResponse_UNKNOWN
	StatusServing        = grpc_health_v1.HealthCheckResponse_SERVING
	StatusNotServing     = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	StatusServiceUnknown = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
)

var (
	// force MServer implements Server.
	_ Server = &MServer{}
)

// NewServer return new gRPC health server.
func NewServer(m map[string]CheckFunc, opts ...ServerOption) *MServer {
	srv := &MServer{
		checks: m,
		Server: health.NewServer(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	// default if not set
	if srv.interval == 0 {
		srv.interval = 60 * time.Second
	}
	if srv.log == nil {
		srv.log = log.Root()
	}
	if srv.timeout == 0 {
		srv.timeout = 2 * time.Second
	}
	srv.ticker = time.NewTicker(srv.interval)

	return srv
}

// Register implements server.Service.
func (s *MServer) Register(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, s)
}

// Init inits health status; start a background job in a separate goroutine to check
// the services' status base on the given interval.
func (s *MServer) Init(status Status) error {
	s.SetServingStatus("", status)
	// if there is no dependent services, don't need to do anything else.
	if len(s.checks) == 0 {
		return nil
	}
	// if there are dependent services, set overall status and all dependent services
	// to NotServing as we don't know their status yet.
	s.SetServingStatus("", StatusNotServing)
	for name := range s.checks {
		s.SetServingStatus(name, StatusNotServing)
	}
	// start a first check immediately.
	s.checkAll()
	// schedule the check
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkAll()
			}
		}
	}()
	return nil
}

func (s *MServer) checkAll() {
	logger := s.log.Fields("correlation_id", uuid.New().String())
	bg := time.Now()
	overall := StatusServing
	for name, check := range s.checks {
		status := StatusServing
		if err := s.check(name, check); err != nil {
			overall = StatusNotServing
			status = StatusNotServing
			logger.Infof("health check failed, service: %s, err: %v", name, err)
		}
		s.SetServingStatus(name, status)
	}
	s.SetServingStatus("", overall)
	logger.Fields("status", overall, "duration", time.Since(bg)).Info("health check completed")
}

func (s *MServer) check(name string, check CheckFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	ch := make(chan error)
	go func() {
		ch <- check(ctx)
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return errors.New("health: check exceed timeout")
	}
}

// ServeHTTP serve health check status via HTTP.
func (s *MServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	services := map[string]Status{}
	check := func(name string) Status {
		rs, err := s.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{
			Service: name,
		})
		if err != nil {
			s.log.Errorf("health check failed, service: %s, err: %v", name, err)
			return StatusNotServing
		}
		return rs.Status
	}
	overall := check("")
	for name := range s.checks {
		ss := check(name)
		services[name] = ss
		// if any error, overall status is set to not serving.
		if ss != StatusServing {
			overall = StatusNotServing
		}
	}
	b, err := json.Marshal(map[string]interface{}{
		"status":   overall,
		"services": services,
	})
	if err != nil {
		b = []byte(fmt.Sprintf(`{"status":%d}`, StatusNotServing))
		s.log.Errorf("failed to marshal data, err: %v", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Close sets all serving status to NOT_SERVING, and configures the server to
// ignore all future status changes.
func (s *MServer) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.Server != nil {
		s.Server.Shutdown()
	}
	return nil
}
