package memory

import (
	"context"
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
	}
	value struct {
		val []byte
		exp time.Time
	}
)

var (
	_ cache.Cacher   = (*Memory)(nil)
	_ health.Checker = (*Memory)(nil)
)

// New return new memory cache.
// Interval is optional, but if set it will be used for cleanup
// the expired keys periodically.
func New(interval ...time.Duration) *Memory {
	t := 500 * time.Millisecond
	if len(interval) > 0 {
		t = interval[0]
	}
	m := &Memory{
		interval: t,
		values:   &sync.Map{},
		exit:     make(chan struct{}),
	}
	return m
}

// Get a value.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
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
	m.values.Delete(key)
	return nil
}

func (m *Memory) clean() {
	tik := time.NewTicker(m.interval)
	defer tik.Stop()
	for {
		select {
		case <-tik.C:
			m.values.Range(func(k, v any) bool {
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
	return nil
}

// Close close underlying resources.
func (m *Memory) Close(ctx context.Context) error {
	close(m.exit)
	return nil
}

// CheckHealth return health check func.
func (m *Memory) CheckHealth(ctx context.Context) error {
	// do nothing.
	return nil
}
