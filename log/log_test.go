package log_test

import (
	"context"
	"os"
	"testing"

	"github.com/pthethanh/micro/log"
)

func TestLog(t *testing.T) {
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_FIELDS", "service:micro,site:vn")

	log.Init(log.FromEnv())

	log.Debug("debug", 1)
	log.Debugf("debug %d", 2)
	log.Info("info", 1)
	log.Infof("info %d", 2)
	log.Warn("warn", 1)
	log.Warnf("warn %d", 2)
	log.Error("error", 1)
	log.Errorf("error %d", 2)
	log.Fields("name", "my application", "address", "1.1.1.1", "instance", 3).Info("application info")
	log.Fields("name", "my application", "address", "1.1.1.1", "application info").Info()

	ctx := log.NewContext(context.Background(), log.Root().Fields("request-id", 123))
	log.Context(ctx).Info("context logger with request id")

	log.Fields("field1", 1).Fields("field2", 2).Context(ctx).Debug()
	log.Fields("field1", 1).Context(ctx).Fields("field2", 2).Debug()
	log.Context(ctx).Fields("field1", 1).Fields("field2", 2).Debug()
}
