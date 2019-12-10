package multilimiter

import (
	"context"
	"sync"
)

// Provides a thread-safe mechanism for acquiring and releasing slots
// Calling Stop() will unblock Acquire() if it's waiting on a slot
type ConcLimiter interface {
	// Wait for a slot to become available
	// only DeadlineExceeded and LimiterStopped errors can be returned
	Acquire(ctx context.Context) (Slot, error)
	// Put the slot back into the pool
	// Release()
	// Cancels processing outstanding acquisition requests
	Cancel()
	// The configured concurrency
	Concurrency() int
	// Wait for all Slots to be returned
	Wait()
}

type Slot interface {
	Release()
}

type slot struct {
	once      sync.Once
	releaseFn func()
}

func (me *slot) Release() {
	me.once.Do(me.releaseFn)
}

type BasicConcLimiter struct {
	size        int
	slots, done chan struct{}
	canceler    *Canceler
	wg          sync.WaitGroup
}

var _ ConcLimiter = (*BasicConcLimiter)(nil)

// Creates a new concurrency limiter
// if size is <= 1, a default of 1 will be used
func NewConcLimiter(size int) *BasicConcLimiter {
	if size <= 1 {
		size = 1
	}

	slots := make(chan struct{}, size)

	for i := 0; i < size; i++ {
		slots <- struct{}{}
	}

	return &BasicConcLimiter{size: size, slots: slots, canceler: NewCanceler()}
}

func (me *BasicConcLimiter) Acquire(ctx context.Context) (Slot, error) {
	if me.canceler.IsCanceled() {
		return nil, LimiterStopped
	}

	// wait for a slot to become available
	select {
	case <-me.canceler.Done():
		return nil, LimiterStopped
	case <-ctx.Done():
		return nil, DeadlineExceeded
	case <-me.slots:
		me.wg.Add(1)
		return &slot{releaseFn: me.release}, nil
	}
}

func (me *BasicConcLimiter) Cancel() {
	me.canceler.Cancel()
}

func (me *BasicConcLimiter) release() {
	if me.size >= 1 {
		me.wg.Done()
		me.slots <- struct{}{}
	}
}

func (me *BasicConcLimiter) Concurrency() int {
	return me.size
}

func (me *BasicConcLimiter) Wait() {
	me.wg.Wait()
}
