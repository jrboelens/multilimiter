package multilimiter

import (
	"context"
	"time"

	"github.com/juju/ratelimit"
)

type RateLimiter interface {
	// Wait until there are resources available
	// only DeadlineExceeded and LimiterStopped errors can be returned
	Wait(ctx context.Context) error
	// The configured rate
	Rate() float64
	// Cancels Wait()ing
	Cancel()
}

type TestableRateLimiter interface {
	RateLimiter
	XXX_TEST_Wait(tokens int64, ctx context.Context) error
}

type BasicRateLimiter struct {
	rate     float64
	bucket   *ratelimit.Bucket
	canceler *Canceler
}

var _ RateLimiter = (*BasicRateLimiter)(nil)

// Returns a *BasicRateLimiter if rate > 0; otherwise a *NoLimitRateLimiter
func NewRateLimiter(rate float64) RateLimiter {
	if rate <= 0 {
		return &NoLimitRateLimiter{}
	}
	bucket := ratelimit.NewBucketWithRate(rate, int64(rate))
	return &BasicRateLimiter{rate: rate, bucket: bucket, canceler: NewCanceler()}
}

// This allows us to force a timeout in testing by setting the number of desired tokens to a high value
func (me *BasicRateLimiter) XXX_TEST_Wait(tokens int64, ctx context.Context) error {
	return me.wait(tokens, ctx)
}

func (me *BasicRateLimiter) Wait(ctx context.Context) error {
	return me.wait(1, ctx)
}

func (me *BasicRateLimiter) wait(tokens int64, ctx context.Context) error {
	if me.canceler.IsCanceled() {
		return LimiterStopped
	}

	if d := me.bucket.Take(tokens); d > 0 {
		select {
		case <-me.canceler.Done():
			return LimiterStopped
		case <-ctx.Done():
			return DeadlineExceeded
		case <-time.After(d):
			return nil
		}
	}
	return nil
}

func (me *BasicRateLimiter) Rate() float64 {
	return me.rate
}

func (me *BasicRateLimiter) Cancel() {
	me.canceler.Cancel()
}

// A Null implementation of RateLimiter
type NoLimitRateLimiter struct{}

var _ RateLimiter = (*NoLimitRateLimiter)(nil)

func (me *NoLimitRateLimiter) Wait(ctx context.Context) error {
	return nil
}

func (me *NoLimitRateLimiter) Rate() float64 {
	return 0.0
}

func (me *NoLimitRateLimiter) Cancel() {}
