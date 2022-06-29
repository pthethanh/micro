package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/grpc/test/grpc_testing"
)

var (
	addrs = map[string]bool{":8000": false, ":8001": false, ":8002": false, ":8003": false, ":8004": false, ":8080": false}
	mu    = &sync.Mutex{}
)

const (
	userName = "test"
)

type (
	httpValidationFunc = func(rs *http.Response) error
	testService        struct {
		grpc_testing.UnimplementedTestServiceServer
	}
)

func (srv *testService) UnaryCall(context.Context, *grpc_testing.SimpleRequest) (*grpc_testing.SimpleResponse, error) {
	return &grpc_testing.SimpleResponse{
		Username: userName,
	}, nil
}

func (srv *testService) Register(s *grpc.Server) {
	grpc_testing.RegisterTestServiceServer(s, srv)
}

func TestInitServerDefault(t *testing.T) {
	t.Parallel()
	addr := availableAddress()
	if addr == "" {
		log.Warn("address is already in use, ignore unit test")
		t.SkipNow()
		return
	}
	os.Setenv("ADDRESS", addr)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.ListenAndServeContext(ctx); err != nil {
		if err != ctx.Err() {
			t.Error(err)
		}
	}
}

func TestInitServerWithOptions(t *testing.T) {
	t.Parallel()
	lis := bufconn.Listen(2000)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	srv := server.New(
		server.Listener(lis),
		server.JWT("secret"),
		server.Logger(log.Root()),
		server.HealthCheck("/health", health.NewServer(nil)),
		server.ServeMuxOptions(server.DefaultHeaderMatcher()),
		server.Options(grpc.ConnectionTimeout(20*time.Second)),
		server.Timeout(20*time.Second, 20*time.Second),
	)
	if err := srv.ListenAndServeContext(ctx); err != nil {
		if err != ctx.Err() {
			t.Error(err)
		}
	}
}

func TestServerAPIs(t *testing.T) {
	addr := availableAddress()
	if addr == "" {
		log.Warn("address is already in use, ignore unit test")
		t.SkipNow()
		return
	}
	os.Setenv("ADDRESS", addr)
	os.Setenv("JWT_SECRET", "") // no auth
	svc := &testService{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := server.New(
		server.FromEnv(),
		server.RoutesPrioritization(true),
		server.PrefixHandler("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"name":"api"}`))
		})),
		server.HandlerWithOptions("/api/v1/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		}), server.NewHandlerOptions().Methods(http.MethodGet).Queries("name", "status")),
		server.Handler("/api/v1/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"name":"test"}`))
		}), http.MethodGet),
		server.HandlerFunc("/api/v1/test1", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status":"ok"}`))
		}, http.MethodPost),
	)
	go func() {
		if err := srv.ListenAndServeContext(ctx, svc); err != nil {
			if err != ctx.Err() {
				t.Error(err)
			}
		}
	}()
	// wait for the server to start
	time.Sleep(10 * time.Millisecond)
	host := "http://localhost" + addr
	cases := []struct {
		name         string
		path         string
		method       string
		body         interface{}
		code         int
		response_map map[string]interface{}
		response_str string
	}{
		{
			name:   "health check",
			path:   "/internal/health",
			method: http.MethodGet,
			code:   http.StatusOK,
			response_map: map[string]interface{}{
				"status": 1,
			},
		},
		{
			name:         "metrics",
			path:         "/internal/metrics",
			method:       http.MethodGet,
			code:         http.StatusOK,
			response_str: "grpc_server_handled_total",
		},
		{
			name:   "handler, ok",
			path:   "/api/v1/test",
			method: http.MethodGet,
			code:   http.StatusOK,
			response_map: map[string]interface{}{
				"name": "test",
			},
		},
		{
			name:   "handler func, ok",
			path:   "/api/v1/test1",
			method: http.MethodPost,
			code:   http.StatusCreated,
			response_map: map[string]interface{}{
				"status": "ok",
			},
		},
		{
			name:   "handler with opts, ok",
			path:   "/api/v1/test?name=status",
			method: http.MethodGet,
			code:   http.StatusOK,
			response_map: map[string]interface{}{
				"status": "ok",
			},
		},
		{
			name:   "handler with prefix works OK with route prioritization",
			path:   "/api/v2/test",
			method: http.MethodGet,
			code:   http.StatusOK,
			response_map: map[string]interface{}{
				"name": "api",
			},
		},
		{
			name:   "handler, not found",
			path:   "/v1/test",
			method: http.MethodPost,
			code:   http.StatusNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var f httpValidationFunc
			if c.response_map != nil {
				f = validateJSONResponse(c.response_map)
			} else {
				f = validateStringResponse(c.response_str)
			}
			if err := testHTTP(host+c.path, c.method, c.body, c.code, f); err != nil {
				t.Error(err)
			}
		})
	}

	// test grpc
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	client := grpc_testing.NewTestServiceClient(conn)
	res, err := client.UnaryCall(context.Background(), &grpc_testing.SimpleRequest{
		FillUsername: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Username != userName {
		t.Errorf("got user_name=%s, want user_name=%s", res.Username, userName)
	}
}

func validateJSONResponse(expect map[string]interface{}) func(rs *http.Response) error {
	return func(res *http.Response) error {
		var body map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			return err
		}
		for k, v := range expect {
			v1 := body[k]
			if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", v1) {
				return fmt.Errorf("got response[%s]=%v, want response[%s]=%v", k, v1, k, v)
			}
		}
		return nil
	}
}

func validateStringResponse(sub string) func(rs *http.Response) error {
	return func(res *http.Response) error {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if !strings.Contains(string(b), sub) {
			return fmt.Errorf("got response string=%s, want response string contains %s", string(b), sub)
		}
		return nil
	}
}

func testHTTP(path string, method string, body interface{}, code int, f httpValidationFunc) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	c := &http.Client{
		Timeout: time.Second,
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != code {
		return fmt.Errorf("got status_code=%d, want status_code=%d", res.StatusCode, code)
	}
	if err := f(res); err != nil {
		return err
	}
	return nil
}

func availableAddress() string {
	mu.Lock()
	defer mu.Unlock()
	for addr, used := range addrs {
		if used {
			continue
		}
		addrs[addr] = true
		l, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		l.Close()
		return addr
	}
	return ""
}
