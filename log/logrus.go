package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	// Logrus implement Logger interface using  logrus.
	Logrus struct {
		logger *logrus.Entry
	}
)

// NewLogrus return new logger with context.
func NewLogrus(opts ...Option) (*Logrus, error) {
	l := &Logrus{}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}
	return l, nil
}

// Init init the logger.
func (l *Logrus) Init(opts ...Option) error {
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
	out, err := options.GetWriter()
	if err != nil {
		return err
	}
	logger := logrus.New()
	logger.SetFormatter(f)
	logger.SetLevel(logrus.Level(level))
	logger.SetOutput(out)
	fields := map[string]any{}
	for k, v := range options.Fields {
		fields[k] = v
	}
	l.logger = logrus.NewEntry(logger).WithFields(fields)
	return nil
}

// Info print info
func (l *Logrus) Info(args ...any) {
	l.logger.Infoln(args...)
}

// Debug print debug
func (l *Logrus) Debug(v ...any) {
	l.logger.Debugln(v...)
}

// Warn print warning
func (l *Logrus) Warn(v ...any) {
	l.logger.Warnln(v...)
}

// Error print error
func (l *Logrus) Error(v ...any) {
	l.logger.Errorln(v...)
}

// Panic panic
func (l *Logrus) Panic(v ...any) {
	l.logger.Panicln(v...)
}

// Infof print info with format.
func (l *Logrus) Infof(format string, v ...any) {
	l.logger.Infof(format, v...)
}

// Debugf print debug with format.
func (l *Logrus) Debugf(format string, v ...any) {
	l.logger.Debugf(format, v...)
}

// Warnf print warning with format.
func (l *Logrus) Warnf(format string, v ...any) {
	l.logger.Warnf(format, v...)
}

// Errorf print error with format.
func (l *Logrus) Errorf(format string, v ...any) {
	l.logger.Errorf(format, v...)
}

// Panicf panic with format.
func (l *Logrus) Panicf(format string, v ...any) {
	l.logger.Panicf(format, v...)
}

// Fields return a new logger with fields.
func (l *Logrus) Fields(kv ...any) Logger {
	return &Logrus{
		logger: l.logger.WithFields(logrus.Fields(fields(kv...))),
	}
}

// Context return new logger from context.
func (l *Logrus) Context(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		kv := make([]any, 0)
		for k, v := range l.logger.Data {
			kv = append(kv, k, v)
		}
		return logger.Fields(kv...)
	}
	return l
}
