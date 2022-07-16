package memory

import "time"

// Interval is an option to override the default expired keys clean up interval.
func Interval(d time.Duration) Option {
	return func(m *Memory) {
		m.interval = d
	}
}

// Shard is an option to set number of shard for sharding the keys.
func Shard(shard uint64) Option {
	return func(m *Memory) {
		m.shard = shard
	}
}
