package timeutil_test

import (
	"testing"
	"time"

	"github.com/pthethanh/micro/util/timeutil"
)

func TestRecorder(t *testing.T) {
	r := timeutil.NewRecorder("get_users", "request_id", 123456789, "test_recorder")
	time.Sleep(200 * time.Millisecond)
	r.Done("validate")
	last := r.Last()
	go func() {
		time.Sleep(100 * time.Millisecond)
		r.DoneSince("concurrent", last)
	}()
	time.Sleep(100 * time.Millisecond)
	r.Done("fetch_from_db")
	time.Sleep(300 * time.Millisecond)
	r.Done("apply_template")
	v := r.Info()
	ctx := "get_users"
	if v.Context != ctx {
		t.Errorf("got context=%s, want context=%s", v.Context, ctx)
	}
	d := 600 * time.Millisecond
	// duration should be more than 600, and less than 700 since there's a step run inside goroutine.
	// +50ms for go scheduler to schedule the goroutine.
	if v.Duration < d || v.Duration > d+50*time.Millisecond {
		t.Errorf("got duration=%d, want duration >= %d and duration < 600+ ", v.Duration, d)
	}
	if len(v.Meta) < 2 {
		t.Errorf("got len(meta)=%d, want len(meta)=%d", len(v.Meta), 2)
	}
	if len(v.Spans) < 3 {
		t.Errorf("got len(spans)=%d, want len(spans)=%d", len(v.Spans), 3)
	}
	t.Logf("text value: %s", v)
}
