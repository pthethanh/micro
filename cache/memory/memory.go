package memory

import (
	"context"
	"sync"
	"time"

	"github.com/pthethanh/micro/cache"
)

type (
	// Memory is an implementation of cache.Cacher
	Memory struct {
		values map[string]value
		sync.RWMutex
		exit chan struct{}
	}
	value struct {
		val []byte
		exp time.Time
	}
)

var (
	_ cache.Cacher = &Memory{}
)

// New return new memory cache.
func New() *Memory {
	m := &Memory{
		values: make(map[string]value),
		exit:   make(chan struct{}),
	}
	return m
}

// Get a value.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	if val, ok := m.values[key]; ok {
		// if cleaner has not done its job yet, go ahead to delete
		if val.expired() {
			go func() {
				m.Delete(ctx, key)
			}()
			return nil, cache.ErrNotFound
		}
		return val.val, nil
	}
	return nil, cache.ErrNotFound
}

// Set a value.
func (m *Memory) Set(ctx context.Context, key string, val []byte, opts ...cache.SetOption) error {
	m.Lock()
	defer m.Unlock()
	opt := &cache.SetOptions{}
	opt.Apply(opts...)
	v := value{
		val: val,
	}
	if opt.TTL != 0 {
		v.exp = time.Now().Add(opt.TTL)
	}
	m.values[key] = v
	return nil
}

// Delete a value.
func (m *Memory) Delete(ctx context.Context, key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.values, key)
	return nil
}

func (m *Memory) clean() {
	tik := time.NewTicker(500 * time.Millisecond)
	defer tik.Stop()
	for {
		select {
		case <-tik.C:
			m.Lock()
			for k, v := range m.values {
				if v.expired() {
					delete(m.values, k)
				}
			}
			m.Unlock()
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
