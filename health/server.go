package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/google/uuid"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/util/syncutil"
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
		// Register register the Server with grpc.Server
		Register(srv *grpc.Server)
		// Init inits overall status using empty string as defined in https://github.com/grpc/grpc/blob/master/doc/health-checking.md
		// and do additional initialize if any.
		Init(status Status) error
		// Close close the underlying resources
		Close() error
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
		srv.timeout = 5 * time.Second
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
	// set dependencies service as not serving.
	for name := range s.checks {
		s.SetServingStatus(name, StatusNotServing)
	}
	// set overall status to not serving if there is any dependencies services.
	if len(s.checks) > 0 {
		s.SetServingStatus("", StatusNotServing)
	}
	// start a first check
	s.checkAll()
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
	logger := s.log.Fields("health_check_id", uuid.New().String())
	logger.Fields("phase", "started").Infof("health check started")
	bg := time.Now()
	overall := StatusServing
	for name, check := range s.checks {
		if err := s.checkAndUpdateStatus(name, check); err != nil {
			overall = StatusNotServing
		}
	}
	s.SetServingStatus("", overall)
	logger.Fields("phase", "completed", "duration", time.Since(bg)).Info("health check completed")
}

func (s *MServer) checkAndUpdateStatus(name string, check CheckFunc) error {
	status := StatusServing
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	f := func(ctx context.Context) { check(ctx) }
	if err := syncutil.WaitCtx(ctx, s.timeout, f); err != nil {
		status = StatusNotServing
		return err
	}
	s.SetServingStatus(name, status)
	return nil
}

// ServeHTTP serve health check status via HTTP.
func (s *MServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	detail := map[string]Status{}
	check := func(name string) Status {
		rs, err := s.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{
			Service: name,
		})
		if err != nil {
			s.log.Fields("health_check_service", name).Errorf("check failed, err: %v", err)
			return StatusNotServing
		}
		return rs.Status
	}
	overall := check("")
	for name := range s.checks {
		ss := check(name)
		detail[name] = ss
		// if any error, overall status is set to not serving.
		if ss != StatusServing {
			overall = StatusNotServing
		}
	}

	b, err := json.Marshal(map[string]interface{}{
		"status":   overall,
		"services": detail,
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
	if s.ticker == nil {
		return nil
	}
	s.ticker.Stop()
	s.Server.Shutdown()
	return nil
}
