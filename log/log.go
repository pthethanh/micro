package log

import (
	"context"
	"fmt"
)

type (
	// Logger defines standard operations of a logger.
	Logger interface {
		// Infof print info with format.
		Infof(format string, v ...interface{})

		// Debugf print debug with format.
		Debugf(format string, v ...interface{})

		// Warnf print warning with format.
		Warnf(format string, v ...interface{})

		// Errorf print error with format.
		Errorf(format string, v ...interface{})

		// Panicf panic with format.
		Panicf(format string, v ...interface{})

		// Info print info.
		Info(v ...interface{})

		// Debug print debug.
		Debug(v ...interface{})

		// Warn print warning.
		Warn(v ...interface{})

		// Error print error.
		Error(v ...interface{})

		// Panic panic.
		Panic(v ...interface{})

		// Fields return new logger with the given fields.
		// The kv should be provided as key values pairs where key is a string.
		Fields(kv ...interface{}) Logger

		// Context provide a way to get a context logger,  i.e... with request-id.
		Context(ctx context.Context) Logger
	}

	// context key
	contextKey string
)

const (
	loggerKey contextKey = contextKey("logger_key")
)

// NewContext return a new logger context.
func NewContext(ctx context.Context, logger Logger) context.Context {
	if logger == nil {
		logger = Root()
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext get logger form context.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return Root()
	}
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return Root()
}

func fields(kv ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	ood := len(kv) % 2
	for i := 0; i < len(kv)-ood; i += 2 {
		fields[fmt.Sprintf("%v", kv[i])] = kv[i+1]
	}
	if ood == 1 {
		fields["msg.1"] = fmt.Sprintf("%v", kv[len(kv)-1])
	}
	return fields
}
