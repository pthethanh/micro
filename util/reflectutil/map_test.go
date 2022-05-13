package reflectutil

import (
	"testing"

	"golang.org/x/exp/maps"
)

func TestJSONObject(t *testing.T) {
	type Counter struct {
		Request int
		Success int
		Fail    int
	}
	got, err := ToMap[string, int](Counter{
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
	if !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
