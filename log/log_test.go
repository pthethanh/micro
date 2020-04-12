package log_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

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

func TestLogInit(t *testing.T) {
	errInitFailed := errors.New("err init failed")
	cases := []struct {
		give func() error
		want error
	}{
		{
			give: func() error {
				return log.Init()
			},
			want: nil,
		},
		{
			give: func() error {
				return log.Init(log.FromEnv())
			},
			want: nil,
		},
		{
			give: func() error {
				return log.Init(log.WithFields("name", "my service"))
			},
			want: nil,
		},
		{
			give: func() error {
				return log.Init(log.WithFields())
			},
			want: nil,
		},
		{
			give: func() error {
				return log.Init(log.WithFormat(log.FormatJSON), log.WithLevel(log.LevelInfo), log.WithTimeFormat(time.RFC1123))
			},
			want: nil,
		},
		{
			give: func() error {
				if err := log.Init(log.WithLevel(log.Level(-1))); err != nil {
					return errInitFailed
				}
				return nil
			},
			want: errInitFailed,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			if err := c.give(); err != c.want {
				t.Errorf("got err=%v, want err=%v", err, c.want)
			}
		})
	}

}

func TestLogWriter(t *testing.T) {
	test := func(w io.ReadWriter) {
		if err := log.Init(log.WithWriter(w), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatJSON)); err != nil {
			t.Errorf("init failed: %v\n", err)
		}
		msg := "something"
		log.Debug(msg)
		b, err := ioutil.ReadAll(w)
		if err != nil {
			t.Error(err)
		}
		if !strings.Contains(string(b), msg) {
			t.Errorf("got msg=%s, want msg constains %q", string(b), msg)
		}
	}
	test(&bytes.Buffer{})
}
