package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/micro/cache"
	"github.com/pthethanh/micro/cache/memory"
)

func TestCache(t *testing.T) {
	var m cache.Cacher = memory.New()
	if err := m.Set(context.Background(), "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if v, err := m.Get(context.Background(), "k"); err != nil || string(v) != "v" {
		t.Fatalf("got result=%v, err=%v, want result=%v, err=%v", string(v), err, "v", nil)
	}

	if _, err := m.Get(context.Background(), "k-notfound"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}

func TestCacheTimeout(t *testing.T) {
	var m cache.Cacher = memory.New()
	// not ok
	if err := m.Set(context.Background(), "k", []byte("v"), cache.TTL(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(600 * time.Millisecond)
	if _, err := m.Get(context.Background(), "k"); err == nil {
		t.Fatal("got key found, want key not found")
	}

	// ok
	if err := m.Set(context.Background(), "k", []byte("v"), cache.TTL(1000*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if _, err := m.Get(context.Background(), "k"); err != nil {
		t.Fatalf("got err=%v, want err=nil", err)
	}
}

func TestCacheDelete(t *testing.T) {
	var m cache.Cacher = memory.New()
	if err := m.Set(context.Background(), "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete(context.Background(), "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(context.Background(), "k"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}
