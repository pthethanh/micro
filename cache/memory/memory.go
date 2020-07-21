package memory

import (
	"context"
	"sync"
	"time"

	"github.com/pthethanh/micro/cache"
)

type (
	Memory struct {
		values map[string]value
		sync.RWMutex
	}
	value struct {
		val interface{}
		exp time.Time
	}
)

var (
	_ cache.Cacher = New()
)

// New return new memory cache.
func New() *Memory {
	m := &Memory{
		values: make(map[string]value),
	}
	go m.clean()
	return m
}

// Get a value
func (m *Memory) Get(ctx context.Context, key string) (interface{}, error) {
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

// Set a value
func (m *Memory) Set(ctx context.Context, key string, val interface{}, opts ...cache.SetOption) error {
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

// Delete a value
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
		<-tik.C
		m.Lock()
		for k, v := range m.values {
			if v.expired() {
				delete(m.values, k)
			}
		}
		m.Unlock()
	}
}

func (v value) expired() bool {
	if v.exp.IsZero() {
		return false
	}
	return time.Now().After(v.exp)
}
