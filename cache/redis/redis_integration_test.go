//+build integration_test

package redis_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	rdis "github.com/go-redis/redis/v8"
	"github.com/pthethanh/micro/cache"
	"github.com/pthethanh/micro/cache/redis"
)

func TestCache(t *testing.T) {
	var m cache.Cacher = redis.New(&rdis.UniversalOptions{})
	// string
	if err := m.Set(context.Background(), "k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if v, err := m.Get(context.Background(), "k"); err != nil || string(v) != "v" {
		t.Fatalf("got result=%v, err=%v, want result=%v, err=%v", string(v), err, "v", nil)
	}

	//struct
	obj := struct {
		Name string
		Age  int
	}{
		Name: "jack",
		Age:  25,
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Set(context.Background(), "k", b); err != nil {
		t.Fatal(err)
	}
	v, err := m.Get(context.Background(), "k")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(v, &obj); err != nil {
		t.Fatal(err)
	}
	if obj.Name != "jack" || obj.Age != 25 {
		t.Fatalf("got name=%s, age=%d, want name=%s, age=%d", obj.Name, obj.Age, "jack", 25)
	}

	// not found
	if _, err := m.Get(context.Background(), "k-notfound"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}

func TestCacheTimeout(t *testing.T) {
	var m cache.Cacher = redis.New(&rdis.UniversalOptions{})
	// not ok
	if err := m.Set(context.Background(), "k2", []byte("v"), cache.TTL(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(600 * time.Millisecond)
	if _, err := m.Get(context.Background(), "k2"); err == nil {
		t.Fatal("got key found, want key not found")
	}

	// ok
	if err := m.Set(context.Background(), "k2", []byte("v"), cache.TTL(1000*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if _, err := m.Get(context.Background(), "k2"); err != nil {
		t.Fatalf("got err=%v, want err=nil", err)
	}
}

func TestCacheDelete(t *testing.T) {
	var m cache.Cacher = redis.New(&rdis.UniversalOptions{})
	if err := m.Set(context.Background(), "k3", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete(context.Background(), "k3"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(context.Background(), "k3"); err != cache.ErrNotFound {
		t.Fatalf("got err=%v, want err=%v", err, cache.ErrNotFound)
	}
}
