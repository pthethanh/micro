package syncutil

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrTimeout timeout exceed error.
	ErrTimeout = errors.New("timeutil: timeout exceed")
)

// WaitCtx executes the given jobs in goroutines, and wait until the jobs done
// or context deadline exceed. DefaultTimeout will be applied if context doesn't has deadline.
// Error if deadline exceed.
func WaitCtx(ctx context.Context, defaultTimeout time.Duration, jobs ...func(ctx context.Context)) error {
	var timeoutCtx = ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	for i := 0; i < len(jobs); i++ {
		go func(i int) {
			defer wg.Done()
			jobs[i](timeoutCtx)
		}(i)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return ErrTimeout
	}
}
