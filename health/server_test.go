package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/pthethanh/micro/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestHealthCheck(t *testing.T) {
	// init, make service 1 ok, service 2 and 3 not ok
	errSrv := errors.New("down")
	m := sync.Map{}
	m.Store("error", errSrv)
	m.Store("sleep", int(100*time.Second))
	srv := health.NewServer(map[string]health.Checker{
		"pkg.v1.MyService1": health.CheckFunc(func(ctx context.Context) error {
			// ok
			return nil
		}),
		"pkg.v1.MyService2": health.CheckFunc(func(ctx context.Context) error {
			d, _ := m.Load("sleep")
			time.Sleep(time.Duration(d.(int)))
			return nil
		}),
		"pkg.v1.MyService3": health.CheckFunc(func(ctx context.Context) error {
			err, _ := m.Load("error")
			if err == nil {
				return nil
			}
			return err.(error)
		}),
	}, health.Interval(500*time.Millisecond), health.Timeout(200*time.Millisecond))
	srv.Init(health.StatusServing)
	defer srv.Close()
	// start a gRPC server.
	lis := bufconn.Listen(2000)
	server := grpc.NewServer()
	defer server.Stop()
	defer lis.Close()
	srv.Register(server)
	go func() {
		if err := server.Serve(lis); err != nil && err != net.ErrClosed {
			panic(err)
		}
	}()

	testGRPC := func(name string, expect health.Status) {
		dialer := func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}
		conn, err := grpc.Dial("", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		rs, err := health.NewClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
			Service: name,
		})
		if err != nil {
			t.Fatal(err)
		}
		if rs.Status != expect {
			name := name
			if name == "" {
				name = "overall"
			}
			t.Fatalf("got %s status=%v, want status=%v", name, rs.Status, expect)
		}
	}

	testHTTP := func(expect health.Status) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("", http.MethodGet, nil)
		if err != nil {
			t.Fatal(err)
		}
		srv.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("got status_code=%d, want status_code=%d", w.Code, http.StatusOK)
		}
		var m map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		status, ok := m["status"]
		if !ok {
			t.Fatalf("got status=empty, want status=%d", expect)
		}
		if v, ok := status.(float64); !ok || health.Status(v) != expect {
			t.Fatalf("got status=%v, want status=%d", status, expect)
		}
	}
	// first check should be not ok, since service 2, 3 are not ok.
	testGRPC("", health.StatusNotServing)
	testGRPC("pkg.v1.MyService1", health.StatusServing)
	testGRPC("pkg.v1.MyService2", health.StatusNotServing)
	testGRPC("pkg.v1.MyService3", health.StatusNotServing)
	testHTTP(health.StatusNotServing)
	// make all services ok.
	m.Store("sleep", 0)
	m.Store("error", nil)
	// wait for next circle before checking
	time.Sleep(500 * time.Millisecond)
	testGRPC("", health.StatusServing)
	testGRPC("pkg.v1.MyService1", health.StatusServing)
	testGRPC("pkg.v1.MyService2", health.StatusServing)
	testGRPC("pkg.v1.MyService3", health.StatusServing)
	testHTTP(health.StatusServing)
}
