package multilimiter_test

import (
	"testing"
	"time"

	"github.com/jrboelens/multilimiter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRateLimiterSpec(t *testing.T) {

	DEFAULT_RATE := 100.0

	Convey("RateLimiter tests ", t, func() {

		Convey("Cancel is idempotent", func() {
			lim := multilimiter.NewRateLimiter(DEFAULT_RATE)
			lim.Cancel()
			So(lim.Cancel, ShouldNotPanic)
		})

		Convey("Wait can be cancelled", func() {
			lim := multilimiter.NewRateLimiter(DEFAULT_RATE)
			timeout := time.Millisecond * 100

			err := lim.Wait(Context(timeout))
			So(err, ShouldBeNil)

			lim.Cancel()

			err = lim.Wait(Context(timeout))
			So(err, ShouldEqual, multilimiter.LimiterStopped)
		})

		Convey("Wait adheres to a timeout", func() {
			timeout := 20 * time.Millisecond
			lim := multilimiter.NewRateLimiter(1.0)

			// 100 tokens represents 100 seconds based on how the token bucket limiter
			// is wired into the BasicRateLimiter
			// normally it waits on 1 token per call to Wait()
			// this test forces a long wait on the tokens in order to trigger the timeout
			typedLim := lim.(multilimiter.TestableRateLimiter)
			err := typedLim.XXX_TEST_Wait(100, Context(timeout))
			So(err, ShouldEqual, multilimiter.DeadlineExceeded)
		})

		Convey("Wait allows a 0 timeout in the context", func() {
			lim := multilimiter.NewRateLimiter(DEFAULT_RATE)
			err := lim.Wait(Context(time.Second * 0))
			So(err, ShouldBeNil)
		})

		Convey("a rate of zero means no limit, not do nothing", func() {
			lim := multilimiter.NewRateLimiter(0.0)
			err := lim.Wait(Context(time.Millisecond * 100))
			So(err, ShouldBeNil)
		})

		Convey("a rate of less than zero means no limit, not do nothing", func() {
			lim := multilimiter.NewRateLimiter(-100.0)
			err := lim.Wait(Context(time.Millisecond * 100))
			So(err, ShouldBeNil)
		})
	})
}
