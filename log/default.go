package log

import "context"

var (
	root Logger
)

func init() {
	if root == nil {
		if err := Init(FromEnv()); err != nil {
			panic(err)
		}
	}
}

// Root return default logger instance.
// Root will try to init a root logger base on environemnt configuration.
// It will panic if failed to init.
func Root() Logger {
	if root == nil {
		if err := Init(FromEnv()); err != nil {
			panic(err)
		}
	}
	return root
}

// Init init the root logger with fields.
func Init(opts ...Option) error {
	l := &microLogger{}
	if err := l.Init(opts...); err != nil {
		return err
	}
	root = l
	return nil
}

// Infof print info with format.
func Infof(format string, v ...interface{}) {
	Root().Infof(format, v...)
}

// Debugf print debug with format.
func Debugf(format string, v ...interface{}) {
	Root().Debugf(format, v...)
}

// Warnf print warning with format.
func Warnf(format string, v ...interface{}) {
	Root().Warnf(format, v...)
}

// Errorf print error with format.
func Errorf(format string, v ...interface{}) {
	Root().Errorf(format, v...)
}

// Panicf panic with format.
func Panicf(format string, v ...interface{}) {
	Root().Panicf(format, v...)
}

// Info print info.
func Info(v ...interface{}) {
	Root().Info(v...)
}

// Debug print debug.
func Debug(v ...interface{}) {
	Root().Debug(v...)
}

// Warn print warning.
func Warn(v ...interface{}) {
	Root().Warn(v...)
}

// Error print error.
func Error(v ...interface{}) {
	Root().Error(v...)
}

// Panic panic.
func Panic(v ...interface{}) {
	Root().Panic(v...)
}

// Fields return a new logger entry with fields.
func Fields(kv ...interface{}) Logger {
	return Root().Fields(kv...)
}

// Context return a logger from the given context.
func Context(ctx context.Context) Logger {
	return FromContext(ctx)
}
