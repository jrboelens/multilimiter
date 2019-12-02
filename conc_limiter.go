package multilimiter

import (
	"context"
)

// Provides a thread-safe mechanism for acquiring and releasing slots
// Calling Stop() will unblock Acquire() if it's waiting on a slot
type ConcLimiter interface {
	// Wait for a slot to become available
	// only DeadlineExceeded and LimiterStopped errors can be returned
	Acquire(ctx context.Context) error
	// Put the slot back into the pool
	Release()
	// Cancels processing outstanding acquisition requests
	Cancel()
	// The configured concurrency
	Concurrency() int
}

type BasicConcLimiter struct {
	size        int
	slots, done chan struct{}
	canceler    *Canceler
}

var _ ConcLimiter = (*BasicConcLimiter)(nil)

// Creates a new concurrency limiter
// if size is <= 1, a default of 1 will be used
func NewConcLimiter(size int) *BasicConcLimiter {
	if size <= 1 {
		size = 1
	}

	slots := make(chan struct{}, size)
	done := make(chan struct{})

	for i := 0; i < size; i++ {
		slots <- struct{}{}
	}

	return &BasicConcLimiter{size: size, slots: slots, done: done, canceler: &Canceler{}}
}

func (me *BasicConcLimiter) Acquire(ctx context.Context) error {
	if me.canceler.IsCanceled() {
		return LimiterStopped
	}

	// wait for a slot to become available
	select {
	case <-me.done:
		return LimiterStopped
	case <-ctx.Done():
		return DeadlineExceeded
	case <-me.slots:
		return nil
	}
}

func (me *BasicConcLimiter) Cancel() {
	me.canceler.Cancel(func() {
		close(me.done)
	})
}

func (me *BasicConcLimiter) Release() {
	if me.size >= 1 {
		me.slots <- struct{}{}
	}
}

func (me *BasicConcLimiter) Concurrency() int {
	return me.size
}