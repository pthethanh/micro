package cache

import "time"

type (
	// SetOptions hold options when setting value for a key.
	SetOptions struct {
		TTL time.Duration
	}
	// SetOption is option when setting value for a key.
	SetOption func(*SetOptions)

	// Cache is interface for a cache service.
	Cache interface {
		// Get a value, return ErrNotFound if key not found.
		Get(key string) (interface{}, error)
		// Set a value
		Set(key string, val interface{}, opts ...SetOption) error
		// Delete a value
		Delete(key string) error
	}
)

// TTL is an option to set Time To Live for a key.
func TTL(ttl time.Duration) SetOption {
	return func(opts *SetOptions) {
		opts.TTL = ttl
	}
}

func (opt *SetOptions) Apply(opts ...SetOption) {
	for _, op := range opts {
		op(opt)
	}
}
