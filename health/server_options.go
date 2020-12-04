package health

import (
	"time"

	"github.com/pthethanh/micro/log"
)

//Interval is an option to set interval for health check.
func Interval(d time.Duration) ServerOption {
	return func(srv *MServer) {
		srv.interval = d
	}
}

// Timeout is an option to set timeout for each service health check.
func Timeout(d time.Duration) ServerOption {
	return func(srv *MServer) {
		srv.timeout = d
	}
}

// Logger is an option to set logger for the health check server.
func Logger(l log.Logger) ServerOption {
	return func(srv *MServer) {
		srv.log = l
	}
}
