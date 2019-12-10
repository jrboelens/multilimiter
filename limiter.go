package multilimiter

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

type Limiter interface {
	// Stops the limiter
	Stop()
	// Waits for all executions to complete before returning
	Wait()
	// Once available time or concurrency becomes available
	// execute function fn in a go routine
	// DeadlineExceeded is returned if the timeout elapses before rate and concurrency slots can be acquired
	// fn's implementer can choose whether to adhere to the Context parameter's Doneness
	Execute(ctx context.Context, fn func(context.Context)) error
}

// Ensure the Limiter implementation always meets the MultiLimiter interface
var _ Limiter = (*BasicLimiter)(nil)

// A limiter that supports limiting by concurrency and rate
type BasicLimiter struct {
	allOpts     *options
	concLimiter ConcLimiter
	rateLimiter RateLimiter
	canceler    *Canceler
}

func NewLimiter(opts ...Option) *BasicLimiter {
	// Add validation here

	allOpts := CreateOptions(opts...)

	return &BasicLimiter{
		allOpts:     allOpts,
		concLimiter: allOpts.concLimit.Limiter,
		rateLimiter: allOpts.rateLimit.Limiter,
		canceler:    NewCanceler(),
	}
}

func DefaultLimiter(rate float64, concurrency int) *BasicLimiter {
	rateOpt := &RateLimitOption{NewRateLimiter(rate)}
	concOpt := &ConcLimitOption{NewConcLimiter(concurrency)}
	return NewLimiter(rateOpt, concOpt)
}

// Stops the limiter
func (me *BasicLimiter) Stop() {
	me.canceler.Cancel()
}

// Waits for all executions to complete before returning
func (me *BasicLimiter) Wait() {
	me.concLimiter.Wait()
}

// Once available time or concurrency becomes available
// execute function fn in a go routine
// DeadlineExceeded is returned if the timeout elapses before rate and concurrency slots can be acquired
func (me *BasicLimiter) Execute(ctx context.Context, fn func(context.Context)) error {
	if me.canceler.IsCanceled() {
		return LimiterStopped
	}

	// wait for a slot from the concurrency pool
	slot, err := me.concLimiter.Acquire(ctx)
	if err != nil {
		return err
	}

	// wait for a token from the rate limiter
	if err := me.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	go func() {
		defer func() {
			r := recover()
			//			me.concLimiter.Release()
			slot.Release()
			if r != nil {
				OutStream.Write([]byte(fmt.Sprintf("Panic found in BasicLimiter: %s\n", r)))
				OutStream.Write(debug.Stack())
				panic(r)
			}
		}()

		fn(ctx)
	}()
	return nil
}

// If Limiter.Execute() panicks the stack trace will be sent down OutStream
// The default value is os.Stdout
var OutStream io.Writer

func init() {
	OutStream = os.Stdout
}
