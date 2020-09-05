package server_test

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	addrs = map[string]bool{":8000": false, ":8001": false, ":8002": false, ":8003": false, ":8004": false, ":8080": false}
	mu    = &sync.Mutex{}
)

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
		server.HealthCheckPaths("/ready", "/live"),
		server.ServeMuxOptions(server.DefaultHeaderMatcher()),
		server.Options(grpc.ConnectionTimeout(20*time.Second)),
		server.Timeout(20*time.Second, 20*time.Second),
		server.HealthChecks(func(ctx context.Context) error {
			return nil
		}),
	)
	if err := srv.ListenAndServeContext(ctx); err != nil {
		if err != ctx.Err() {
			t.Error(err)
		}
	}
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
