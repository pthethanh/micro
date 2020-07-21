package memory_test

import (
	"testing"
	"time"

	"github.com/pthethanh/micro/cache"
	"github.com/pthethanh/micro/cache/memory"
)

func TestCache(t *testing.T) {
	var m cache.Cache = memory.New()
	if err := m.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	if v, err := m.Get("k"); err != nil || v.(string) != "v" {
		t.Fatal(err)
	}
}

func TestCacheTimeout(t *testing.T) {
	var m cache.Cache = memory.New()
	// not ok
	if err := m.Set("k", "v", cache.TTL(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(600 * time.Millisecond)
	if _, err := m.Get("k"); err == nil {
		t.Fatal("got key found, want key not found")
	}

	// ok
	if err := m.Set("k", "v", cache.TTL(1000*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if _, err := m.Get("k"); err != nil {
		t.Fatalf("got err=%v, want err=nil", err)
	}
}

func TestCacheDelete(t *testing.T) {
	var m cache.Cache = memory.New()
	if err := m.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete("k"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get("k"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}
