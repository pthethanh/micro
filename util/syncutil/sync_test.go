package syncutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/micro/util/syncutil"
)

func TestWaitCtx(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		ctx         time.Duration
		def         time.Duration
		jobs        []func(ctx context.Context)
		expectedErr error
	}{
		{
			name: "context no deadline, jobs done before timeout",
			ctx:  0,
			def:  5 * time.Second,
			jobs: []func(ctx context.Context){
				func(context.Context) {
					time.Sleep(100 * time.Microsecond)
				},
			},
			expectedErr: nil,
		},
		{
			name: "context has deadline, default timeout less than deadline, jobs done before timeout",
			ctx:  5 * time.Second,
			def:  100 * time.Microsecond,
			jobs: []func(ctx context.Context){
				func(context.Context) {
					time.Sleep(100 * time.Microsecond)
				},
			},
			expectedErr: nil,
		},
		{
			name: "context has no deadline, jobs takes longer time to complete than default timeout",
			ctx:  0,
			def:  200 * time.Microsecond,
			jobs: []func(ctx context.Context){
				func(context.Context) {
					time.Sleep(1 * time.Second)
				},
			},
			expectedErr: syncutil.ErrTimeout,
		},
		{
			name: "context has deadline, jobs takes longer time to complete than deadline",
			ctx:  200 * time.Microsecond,
			def:  5 * time.Second,
			jobs: []func(ctx context.Context){
				func(context.Context) {
					time.Sleep(1 * time.Second)
				},
			},
			expectedErr: syncutil.ErrTimeout,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.ctx != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, test.ctx)
				defer cancel()
			}
			if err := syncutil.WaitCtx(ctx, test.def, test.jobs...); err != test.expectedErr {
				t.Error(err)
			}
		})
	}
}
