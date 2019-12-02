package multilimiter_test

import (
	"context"
	"time"

	"github.com/jrboelens/multilimiter"
)

func ContextWithCancel(timeout time.Duration) (context.Context, context.CancelFunc) {
	if int64(timeout) <= 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), timeout)
}

func Context(timeout time.Duration) context.Context {
	ctx, _ := ContextWithCancel(timeout)
	return ctx
}

func NewLimiter(rate float64, concurrency int) multilimiter.Limiter {
	return multilimiter.DefaultLimiter(rate, concurrency)
}

func NewDefaultLimiter() multilimiter.Limiter {
	return NewLimiter(DEFAULT_RATE, DEFAULT_CONCURRENCY)
}

type SlowRateLimiter struct {
	multilimiter.TestableRateLimiter
	tokens int64
}

func (me *SlowRateLimiter) Wait(ctx context.Context) error {
	return me.TestableRateLimiter.XXX_TEST_Wait(me.tokens, ctx)
}

// This is just here to satisfy the interface
func (me *SlowRateLimiter) XXX_TEST_Wait(tokens int64, ctx context.Context) error {
	return me.TestableRateLimiter.XXX_TEST_Wait(tokens, ctx)
}
