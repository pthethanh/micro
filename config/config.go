package config

type (
	// ReadOptions contains available options of Reader interface.
	ReadOptions struct {
		Prefix string
		File   string
	}

	// ReadOption is a helper setting ReadOptions.
	ReadOption func(o *ReadOptions)

	// Reader is a configuration loader.
	Reader interface {
		// Read read the configuration into the given struct (ptr).
		// The provided struct should be a pointer.
		Read(ptr interface{}, options ...ReadOption) error

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
func WithFile(f string) ReadOption {
	return func(o *ReadOptions) {
		o.File = f
	}
}

// Apply applies the given option.
func (op *ReadOptions) Apply(opts ...ReadOption) {
	for _, opt := range opts {
		opt(op)
	}
}
