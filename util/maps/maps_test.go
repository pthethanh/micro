package maps_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/pthethanh/micro/util/maps"
)

func TestMap(t *testing.T) {
	type Counter struct {
		Request int
		Success int
		Fail    int
	}
	// new
	got, err := maps.New[string, int](Counter{
		Request: 5,
		Success: 4,
		Fail:    1,
	})
	if err != nil {
		t.Error(err)
	}
	want := map[string]int{
		"Request": 5,
		"Success": 4,
		"Fail":    1,
	}
	if !got.Equal(want) {
		t.Errorf("equal test, got %v, want %v", got, want)
	}
	// clone, transform & delete
	got = got.Clone().Filter(func(k string, v int) (bool, string, int) {
		return true, strings.ToLower(k), v * 2
	}).Delete("fail")
	want = map[string]int{
		"success": 8,
		"request": 10,
	}
	if !got.Equal(want) {
		t.Errorf("clone, filter & delete test, got %v, want %v", got, want)
	}
	if len(got.Keys()) != 2 {
		t.Errorf("keys test, got len=%d, want len=2", len(got.Keys()))
	}
	if len(got.Values()) != 2 {
		t.Errorf("values test, got len=%d, want len=2", len(got.Values()))
	}
	// keys
	gotKeys := got.Keys()
	sort.Strings(gotKeys)
	wantKeys := []string{"request", "success"}
	if gotKeys[0] != wantKeys[0] || gotKeys[1] != wantKeys[1] {
		t.Errorf("keys, values test, got keys=%v, want keys=%v", gotKeys, wantKeys)
	}
	// values
	gotValues := got.Values()
	sort.Ints(gotValues)
	wantValues := []int{8, 10}
	if gotValues[0] != wantValues[0] || gotValues[1] != wantValues[1] {
		t.Errorf("keys, values test, got keys=%v, want keys=%v", gotValues, wantValues)
	}
	// copy
	got = map[string]int{
		"request": 10,
		"success": 8,
	}
	m := maps.Map[string, int](map[string]int{
		"fail":  1,
		"total": 19,
	})
	m.Copy(got)
	want = maps.Map[string, int]{
		"request": 10,
		"success": 8,
		"fail":    1, // overridden
		"total":   19,
	}
	if !got.Equal(want) {
		t.Errorf("copy test, got %v, want %v", got, want)
	}
	// set
	got.Set(0)
	want = maps.Map[string, int]{
		"request": 0,
		"success": 0,
		"fail":    0,
		"total":   0,
	}
	if !got.Equal(want) {
		t.Errorf("set test, got %v, want %v", got, want)
	}
	// set some keys
	got.Set(1, "total")
	want = maps.Map[string, int]{
		"request": 0,
		"success": 0,
		"fail":    0,
		"total":   1,
	}
	if !got.Equal(want) {
		t.Errorf("set some keys test, got %v, want %v", got, want)
	}
	// clear
	if len(got.Clear()) != 0 {
		t.Errorf("clear test, got len=%d, want len=0", len(got))
	}
	// equals
	if got.Equal(maps.Map[string, int]{"total": 1}) {
		t.Errorf("equal test, got equal=true, want equal=false")
	}
	if (maps.Map[string, int]{"total": 1}).Equal(maps.Map[string, int]{"total": 2}) {
		t.Errorf("equal value test, got equal=true, want equal=false")
	}
	// new error
	if _, err := maps.New[string, int](make(chan int)); err == nil {
		t.Errorf("new with invalid struct, got err=nil, want err!=nil")
	}
	if _, err = maps.New[string, int](map[int]string{1: "1"}); err == nil {
		t.Errorf("new with invalid map, got err=nil, want err=%v", err)
	}

}
