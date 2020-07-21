package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pthethanh/micro/cache"
)

type (
	// Redis is an implementation of cache.Cacher using Redis.
	Redis struct {
		conn redis.UniversalClient
	}
)

// New return a cacher using Redis.
func New(opts *redis.UniversalOptions) *Redis {
	return &Redis{
		conn: redis.NewUniversalClient(opts),
	}
}

// Get a value, return cache.ErrNotFound if key not found.
func (r *Redis) Get(ctx context.Context, key string) (interface{}, error) {
	cmd := r.conn.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	// TODO implement  me
	return nil, nil
}

// Set a value
func (r *Redis) Set(ctx context.Context, key string, val interface{}, opts ...cache.SetOption) error {
	opt := &cache.SetOptions{}
	opt.Apply(opts...)
	if cmd := r.conn.Set(ctx, key, val, opt.TTL); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

// Delete a value
func (r *Redis) Delete(ctx context.Context, key string) error {
	if cmd := r.conn.Del(ctx, key); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
