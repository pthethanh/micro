package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	microLogger struct {
		logger *logrus.Entry
	}
)

// NewMicroLogger return new logger with context.
func NewMicroLogger(opts ...Option) (Logger, error) {
	l := &microLogger{}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}
	return l, nil
}

// Init init the logger.
func (l *microLogger) Init(opts ...Option) error {
	var f logrus.Formatter
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	format, err := options.GetFormat()
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		f = &logrus.JSONFormatter{
			TimestampFormat: options.TimeFormat,
		}
	case FormatText:
		f = &logrus.TextFormatter{
			TimestampFormat: time.RFC1123,
			FullTimestamp:   true,
		}
	}
	level, err := options.GetLevel()
	if err != nil {
		return err
	}
	out, err := options.GetOutput()
	if err != nil {
		return err
	}
	logger := logrus.New()
	logger.SetFormatter(f)
	logger.SetLevel(logrus.Level(level))
	logger.SetOutput(out)
	fields := map[string]interface{}{}
	for k, v := range options.Fields {
		fields[k] = v
	}
	l.logger = logrus.NewEntry(logger).WithFields(fields)
	return nil
}

// Info print info
func (l *microLogger) Info(args ...interface{}) {
	l.logger.Infoln(args...)
}

// Debugf print debug
func (l *microLogger) Debug(v ...interface{}) {
	l.logger.Debugln(v...)
}

// Warn print warning
func (l *microLogger) Warn(v ...interface{}) {
	l.logger.Warnln(v...)
}

// Errorf print error
func (l *microLogger) Error(v ...interface{}) {
	l.logger.Errorln(v...)
}

// Panic panic
func (l *microLogger) Panic(v ...interface{}) {
	l.logger.Panicln(v...)
}

// Infof print info with format.
func (l *microLogger) Infof(format string, v ...interface{}) {
	l.logger.Infof(format, v...)
}

// Debugf print debug with format.
func (l *microLogger) Debugf(format string, v ...interface{}) {
	l.logger.Debugf(format, v...)
}

// Warnf print warning with format.
func (l *microLogger) Warnf(format string, v ...interface{}) {
	l.logger.Warnf(format, v...)
}

// Errorf print error with format.
func (l *microLogger) Errorf(format string, v ...interface{}) {
	l.logger.Errorf(format, v...)
}

// Panicf panic with format.
func (l *microLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}

// WithFields return a new logger with fields.
func (l *microLogger) Fields(kv ...interface{}) Logger {
	return &microLogger{
		logger: l.logger.WithFields(logrus.Fields(fields(kv...))),
	}
}

// Context return new logger from context.
func (l *microLogger) Context(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		kv := make([]interface{}, 0)
		for k, v := range l.logger.Data {
			kv = append(kv, k, v)
		}
		return logger.Fields(kv...)
	}
	return l
}
