// Package log provides an easy-to-use and structured logger.
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

type (
	// Logger defines standard operations of a logger.
	Logger interface {
		// Init init the logger.
		Init(...Option) error

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

	// Options hold logger options
	Options struct {
		Level      Level             `envconfig:"LOG_LEVEL" default:"5"`
		Format     Format            `envconfig:"LOG_FORMAT" default:"json"`
		TimeFormat string            `envconfig:"LOG_TIME_FORMAT" default:"Mon, 02 Jan 2006 15:04:05 -0700"`
		Output     string            `envconfig:"LOG_OUTPUT"`
		Fields     map[string]string `envconfig:"LOG_FIELDS"`
		writer     io.Writer
	}
	// Option is an option for configure logger.
	Option = func(*Options)

	// Level is log level.
	Level int32

	// Format is log format
	Format string
)

const (
	loggerKey  contextKey = contextKey("logger_key")
	filePrefix            = "file://"
)

// These are the different logging levels.
const (
	// LevelPanic level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	LevelPanic Level = iota
	// LevelFatal level. Logs and then calls os.Exit. It will exit even if the
	// logging level is set to Panic.
	LevelFatal
	// LevelError level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LevelError
	// LevelWarn level. Non-critical entries that deserve eyes.
	LevelWarn
	// LevelInfo level. General operational entries about what's going on inside the
	// application.
	LevelInfo
	// LevelDebug level. Usually only enabled when debugging. Very verbose logging.
	LevelDebug
	// LevelTrace level. Designates finer-grained informational events than the Debug.
	LevelTrace
)

// Formats of log output.
const (
	FormatJSON Format = "json"
	FormatText Format = "text"
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

// GetWriter return writer output. If the given output is not valid, os.Stdout is returned.
func (opts Options) GetWriter() (io.Writer, error) {
	switch {
	case opts.writer != nil:
		return opts.writer, nil
	case strings.HasPrefix(opts.Output, filePrefix):
		name := opts.Output[len(filePrefix):]
		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		return f, nil
	case opts.Output == "":
		return os.Stdout, nil
	default:
		return nil, fmt.Errorf("log: output not supported: %s", opts.Output)
	}
}

// GetLevel return log level. If the given level is not valid, LevelDebug is returned.
func (opts Options) GetLevel() (Level, error) {
	if opts.Level < LevelPanic || opts.Level > LevelTrace {
		return LevelDebug, fmt.Errorf("log: level not supported: %d", opts.Level)
	}
	return opts.Level, nil
}

// GetFormat return format of output log. If given format is not valid, JSON format is returned.
func (opts Options) GetFormat() (Format, error) {
	if opts.Format != FormatText && opts.Format != FormatJSON && opts.Format != "" {
		return "", fmt.Errorf("log: format not supported: %s", opts.Format)
	}
	if opts.Format == "" {
		return FormatJSON, nil
	}
	return opts.Format, nil
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
