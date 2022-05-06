// Package config defines standard interfaces for a config reader/writer.
package config

type (
	// ReadOptions contains available options of Reader interface.
	ReadOptions struct {
		Prefix string
		File   string
		// Same as File but ignore error.
		FileNoErr string
	}

	// ReadOption is a helper setting ReadOptions.
	ReadOption func(o *ReadOptions)

	// Reader is a configuration loader.
	Reader interface {
		// Read read the configuration into the given struct (ptr).
		// The provided struct should be a pointer.
		Read(ptr any, options ...ReadOption) error

		// Close close the underlying source.
		Close() error
	}
)

// WithPrefix return a with prefix reader option.
func WithPrefix(prefix string) ReadOption {
	return func(o *ReadOptions) {
		o.Prefix = prefix
	}
}

// WithFile is an option allow the reader read configuration from a file.
// If the file is not found, error will return.
func WithFile(f string) ReadOption {
	return func(o *ReadOptions) {
		o.File = f
	}
}

// WithFileNoError same as WithFile but ignore error while reading the file.
func WithFileNoError(f string) ReadOption {
	return func(o *ReadOptions) {
		o.FileNoErr = f
	}
}

// Apply applies the given option.
func (op *ReadOptions) Apply(opts ...ReadOption) {
	for _, opt := range opts {
		opt(op)
	}
}
