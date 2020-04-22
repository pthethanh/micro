package log

import (
	"fmt"
	"io"
	"log"

	"github.com/pthethanh/micro/config"
	"github.com/pthethanh/micro/config/envconfig"
)

// FromEnv provides an option to load all options from environment variables.
// LOG_LEVEL default:"5" which is debug level
// LOG_FORMAT default:"json"
// LOG_TIME_FORMAT default:"Mon, 02 Jan 2006 15:04:05 -0700"
// LOG_OUTPUT, default to be stdout, use file://my.log for writing to a file.
// LOG_FIELDS is a map of key/value. i.e: name:myservice,site:vietnam
func FromEnv(readOpts ...config.ReadOption) Option {
	v := &Options{}
	if err := envconfig.Read(v, readOpts...); err != nil {
		log.Println("[ERROR] log: failed to read log environment config, err:", err)
	}
	return FromOptions(v)
}

// FromOptions is an option to create new logger from an existing Options.
func FromOptions(opts *Options) Option {
	return func(v *Options) {
		v.Fields = opts.Fields
		v.Format = opts.Format
		v.Level = opts.Level
		v.TimeFormat = opts.TimeFormat
	}
}

// WithLevel provides an option to set log level.
func WithLevel(level Level) Option {
	return func(opts *Options) {
		opts.Level = level
	}
}

// WithFormat provides an option to set log format.
func WithFormat(f Format) Option {
	return func(opts *Options) {
		opts.Format = f
	}
}

// WithTimeFormat provides an option to set time format for logger.
func WithTimeFormat(f string) Option {
	return func(opts *Options) {
		opts.TimeFormat = f
	}
}

// WithFields provides an option to set logger fields.
func WithFields(kv ...interface{}) Option {
	return func(opts *Options) {
		if opts.Fields == nil {
			opts.Fields = make(map[string]string)
		}
		for k, v := range fields(kv...) {
			opts.Fields[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
		}
	}
}

// WithWriter provides an option to set a output writer.
func WithWriter(w io.Writer) Option {
	return func(opts *Options) {
		opts.writer = w
	}
}
