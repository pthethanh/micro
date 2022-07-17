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
		values   map[int]*sync.Map
		exit     chan struct{}
		opened   bool
		shard    uint64
	}
	value struct {
		val []byte
		exp *time.Time
	}

	Option func(*Memory)
)

var (
	_ cache.Cacher   = (*Memory)(nil)
	_ health.Checker = (*Memory)(nil)

	// ErrInvalidConnectionState indicate that the connection has not been opened properly.
	ErrInvalidConnectionState = errors.New("invalid connection state")
)

const (
	// offset64 FNVa offset basis. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	offset64 uint64 = 14695981039346656037
	// prime64 FNVa prime value. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	prime64 uint64 = 1099511628211
)

// New return new memory cache.
func New(opts ...Option) *Memory {
	m := &Memory{
		interval: 500 * time.Millisecond,
		values:   make(map[int]*sync.Map),
		exit:     make(chan struct{}),
		shard:    10,
	}
	for _, opt := range opts {
		opt(m)
	}
	// init shards
	for i := 0; i < int(m.shard); i++ {
		m.values[i] = &sync.Map{}
	}
	return m
}

// Get a value.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	if !m.opened {
		return nil, ErrInvalidConnectionState
	}
	if v, ok := m.getShard(key).Load(key); ok {
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
		t := time.Now().Add(opt.TTL)
		v.exp = &t
	}
	m.getShard(key).Store(key, v)
	return nil
}

// Delete a value.
func (m *Memory) Delete(ctx context.Context, key string) error {
	if !m.opened {
		return ErrInvalidConnectionState
	}
	m.getShard(key).Delete(key)
	return nil
}

func (m *Memory) clean() {
	tik := time.NewTicker(m.interval)
	defer tik.Stop()
	for i := 0; i < int(m.shard); i++ {
		i := i
		go func() {
			for {
				select {
				case <-tik.C:
					m.values[i].Range(func(k, v interface{}) bool {
						val := v.(value)
						if val.expired() {
							m.values[i].Delete(k)
						}
						return true
					})
				case <-m.exit:
					return
				}
			}
		}()
	}
}

func (v value) expired() bool {
	if v.exp == nil {
		return false
	}
	if v.exp.IsZero() {
		return false
	}
	return time.Now().After(*v.exp)
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

func (m *Memory) getShardIndex(key string) int {
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}
	return int(hash % m.shard)
}

func (m *Memory) getShard(key string) *sync.Map {
	return m.values[m.getShardIndex(key)]
}
