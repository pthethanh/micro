package memory_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pthethanh/micro/cache/memory"
)

func BenchmarkSet(b *testing.B) {
	c := memory.New(memory.Interval(100 * time.Millisecond))
	c.Open(context.Background())
	defer c.Close(context.Background())
	v := []byte("v")
	for i := 0; i < b.N; i++ {
		c.Set(context.Background(), fmt.Sprintf("key-%d", i), v)
	}
}

func BenchmarkGet(b *testing.B) {
	c := memory.New(memory.Interval(100 * time.Millisecond))
	c.Open(context.Background())
	defer c.Close(context.Background())
	v := []byte("v")
	for i := 0; i < 1_000_000; i++ {
		c.Set(context.Background(), fmt.Sprintf("key-%d", i), v)
	}
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c.Get(context.Background(), fmt.Sprintf("key-%d", i))
		}
	})
}
