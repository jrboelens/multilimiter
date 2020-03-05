package multilimiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/jrboelens/multilimiter"
	. "github.com/smartystreets/goconvey/convey"
)

const DEFAULT_RATE = 100.0
const DEFAULT_CONCURRENCY = 1

func TestLimiterSpec(t *testing.T) {

	DEFAULT_CONTEXT := func() context.Context {
		return Context(time.Millisecond * 2000)
	}

	EmptyExecuteFunc := func(context.Context) {}

	Convey("Limiter ", t, func() {

		Convey("walking Skeleton test", func() {
			lim := NewDefaultLimiter()

			lim.Execute(DEFAULT_CONTEXT(), EmptyExecuteFunc)
			lim.Execute(DEFAULT_CONTEXT(), EmptyExecuteFunc)
			err := lim.Execute(DEFAULT_CONTEXT(), EmptyExecuteFunc)
			So(err, ShouldBeNil)

			lim.Wait()
			lim.Stop()
		})

		Convey("Execute can be cancelled", func() {
			lim := NewDefaultLimiter()
			lim.Stop()

			err := lim.Execute(DEFAULT_CONTEXT(), EmptyExecuteFunc)
			So(err, ShouldEqual, multilimiter.LimiterStopped)
		})

		Convey("timeouts occur when", func() {
			Convey("rate limiter cannot acquire rate quickly enough", func() {
				// forcing the rate limit to ask for a large amount of tokens
				// when it first fires up will cause the timeout to occur
				tokens := int64(1000)
				ctx := Context(time.Millisecond * 10)

				rateLim := multilimiter.NewRateLimiter(1)
				typedLim := rateLim.(multilimiter.TestableRateLimiter)
				slowRateLim := &SlowRateLimiter{typedLim, tokens}
				rateOpt := &multilimiter.RateLimitOption{slowRateLim}

				concOpt := &multilimiter.ConcLimitOption{multilimiter.NewConcLimiter(DEFAULT_CONCURRENCY)}

				lim := multilimiter.NewLimiter(rateOpt, concOpt)

				err := lim.Execute(ctx, EmptyExecuteFunc)
				So(err, ShouldEqual, multilimiter.DeadlineExceeded)
			})

			Convey("concurrency limiter cannot acquire a slot quickly enough", func() {
				timeout := time.Millisecond * 100
				lim := NewDefaultLimiter()

				done := make(chan struct{})

				var err1, err2 error
				go func() {
					err1 = lim.Execute(Context(0), func(context.Context) { time.Sleep(timeout * 2) })

					go func() {
						err2 = lim.Execute(Context(timeout), EmptyExecuteFunc)
					}()

					time.Sleep(timeout)
					close(done)
				}()

				<-done

				So(err1, ShouldBeNil)
				So(err2, ShouldEqual, multilimiter.DeadlineExceeded)
			})
		})

		Convey("concurrency", func() {

			// without a sleep we can't guarantee we hit the max concurrency because
			// everything happens so quickly
			DELAY := 75 * time.Millisecond

			RunConcurrencyTest := func(rate float64, concurrency, executions int) {
				fn := func() { time.Sleep(DELAY) }
				lim := NewLimiter(rate, concurrency)
				tracker := ExecutesConcurrently(lim, executions, DEFAULT_CONTEXT(), fn)
				So(tracker.Max(), ShouldEqual, concurrency)
			}

			Convey("only runs one function at a time with 1 concurrency", func() {
				concurrency, executions := 1, 11
				RunConcurrencyTest(DEFAULT_RATE, concurrency, executions)
			})

			Convey("only runs two functions at a time with 2 concurrency", func() {
				concurrency, executions := 2, 11
				RunConcurrencyTest(DEFAULT_RATE, concurrency, executions)
			})

			Convey("only runs 10 functions at a time with 10 concurrency", func() {
				concurrency, executions := 10, 101
				RunConcurrencyTest(1000, concurrency, executions)
			})
		})

		// These tests need to run over a longish interval because the bucket limiter takes time to accrue tokens
		// waiting ~10s makes the limiter around ~85% accurate.  Waiting ~20s makes it around 95% accurate
		Convey("rate", func() {

			RateIsWithinRange := func(rate, minRate, maxRate float64, concurrency, executions int) {
				ctx := context.Background()
				tracker := ExecutesConcurrently(NewLimiter(rate, concurrency), executions, ctx)
				So(tracker.Rate(), ShouldBeBetween, minRate, maxRate)
			}

			Convey("runs at ~1/s", func() {
				rate, minRate, maxRate := 1.0, .5, 1.5
				concurrency, executions := DEFAULT_CONCURRENCY, 12

				RateIsWithinRange(rate, minRate, maxRate, concurrency, executions)
			})

			Convey("runs at ~100/s", func() {
				rate, minRate, maxRate := 100.0, 95.0, 105.0
				concurrency, executions := DEFAULT_CONCURRENCY, 1000

				RateIsWithinRange(rate, minRate, maxRate, concurrency, executions)
			})

			Convey("runs at ~500/s", func() {
				rate, minRate, maxRate := 500.0, 450.0, 550.0
				concurrency, executions := DEFAULT_CONCURRENCY, 1000

				RateIsWithinRange(rate, minRate, maxRate, concurrency, executions)
			})
		})
	})
}

func ExecutesConcurrently(lim multilimiter.Limiter, executions int, ctx context.Context, funcs ...func()) *multilimiter.ConcurrencyTracker {
	tracker := &multilimiter.ConcurrencyTracker{}
	tracker.Start()

	for i := 0; i < executions; i++ {
		lim.Execute(ctx, func(context.Context) {
			tracker.Add()

			for _, fn := range funcs {
				fn()
			}

			tracker.Subtract()
		})
	}
	lim.Wait()
	tracker.Stop()

	So(tracker.Total(), ShouldEqual, executions)
	So(tracker.Current(), ShouldEqual, 0)

	return tracker
}
