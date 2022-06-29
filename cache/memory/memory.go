package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pthethanh/micro/cache"
	"github.com/pthethanh/micro/health"
)

type (
	// Memory is an implementation of cache.Cacher
	Memory struct {
		interval time.Duration
		values   *sync.Map
		exit     chan struct{}
		opened   bool
	}
	value struct {
		val []byte
		exp time.Time
	}

	Option func(*Memory)
)

var (
	_ cache.Cacher   = (*Memory)(nil)
	_ health.Checker = (*Memory)(nil)

	// ErrInvalidConnectionState indicate that the connection has not been opened properly.
	ErrInvalidConnectionState = errors.New("invalid connection state")
)

// New return new memory cache.
func New(opts ...Option) *Memory {
	m := &Memory{
		interval: 500 * time.Millisecond,
		values:   &sync.Map{},
		exit:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Get a value.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	if !m.opened {
		return nil, ErrInvalidConnectionState
	}
	if v, ok := m.values.Load(key); ok {
		val := v.(value)
		// if cleaner has not done its job yet, go ahead to delete
		if val.expired() {
			_ = m.Delete(ctx, key)
			return nil, cache.ErrNotFound
		}
		return val.val, nil
	}
	return nil, cache.ErrNotFound
}

// Set a value.
func (m *Memory) Set(ctx context.Context, key string, val []byte, opts ...cache.SetOption) error {
	if !m.opened {
		return ErrInvalidConnectionState
	}
	opt := &cache.SetOptions{}
	opt.Apply(opts...)
	v := value{
		val: val,
	}
	if opt.TTL != 0 {
		v.exp = time.Now().Add(opt.TTL)
	}
	m.values.Store(key, v)
	return nil
}

// Delete a value.
func (m *Memory) Delete(ctx context.Context, key string) error {
	if !m.opened {
		return ErrInvalidConnectionState
	}
	m.values.Delete(key)
	return nil
}

func (m *Memory) clean() {
	tik := time.NewTicker(m.interval)
	defer tik.Stop()
	for {
		select {
		case <-tik.C:
			m.values.Range(func(k, v interface{}) bool {
				val := v.(value)
				if val.expired() {
					m.values.Delete(k)
				}
				return true
			})
		case <-m.exit:
			return
		}
	}
}

func (v value) expired() bool {
	if v.exp.IsZero() {
		return false
	}
	return time.Now().After(v.exp)
}

// Open make the cacher ready for using.
func (m *Memory) Open(ctx context.Context) error {
	go m.clean()
	m.opened = true
	return nil
}

// Close close underlying resources.
func (m *Memory) Close(ctx context.Context) error {
	m.opened = false
	close(m.exit)
	return nil
}

// CheckHealth return health check func.
func (m *Memory) CheckHealth(ctx context.Context) error {
	if !m.opened {
		return ErrInvalidConnectionState
	}
	return nil
}
