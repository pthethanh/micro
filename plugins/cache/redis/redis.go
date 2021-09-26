package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pthethanh/micro/cache"
)

type (
	// Redis is an implementation of cache.Cacher using Redis.
	Redis struct {
		opts *redis.UniversalOptions
		conn redis.UniversalClient
	}
)

var (
	_ cache.Cacher = (*Redis)(nil)
)

// New return a new cacher using Redis.
func New(opts ...Option) *Redis {
	r := &Redis{
		opts: &redis.UniversalOptions{},
	}
	for _, op := range opts {
		op(r)
	}
	return r
}

// Open open connection to the target servers.
func (r *Redis) Open(ctx context.Context) error {
	r.conn = redis.NewUniversalClient(r.opts)
	return nil
}

// Get a value, return cache.ErrNotFound if key not found.
func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	cmd := r.conn.Get(ctx, key)
	if cmd.Err() == redis.Nil {
		return nil, cache.ErrNotFound
	}
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Bytes()
}

// Set a value
func (r *Redis) Set(ctx context.Context, key string, val []byte, opts ...cache.SetOption) error {
	opt := &cache.SetOptions{}
	opt.Apply(opts...)
	if cmd := r.conn.Set(ctx, key, val, opt.TTL); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

// Delete a value
func (r *Redis) Delete(ctx context.Context, key string) error {
	if cmd := r.conn.Del(ctx, key); cmd.Err() != nil && cmd.Err() != redis.Nil {
		return cmd.Err()
	}
	return nil
}

// Close flush and close the underlying connection.
func (r *Redis) Close(ctx context.Context) error {
	rs := r.conn.FlushAll(ctx)
	if err := rs.Err(); err != nil {
		return err
	}
	return r.conn.Close()
}
