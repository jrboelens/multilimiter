package multilimiter_test

import (
	"testing"
	"time"

	"github.com/jrboelens/multilimiter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConcLimiterSpec(t *testing.T) {

	DEFAULT_CONCURRENCY := 10

	Convey("ConcLimiter tests ", t, func() {

		Convey("Cancel is idempotent", func() {
			lim := multilimiter.NewConcLimiter(DEFAULT_CONCURRENCY)
			lim.Cancel()
			So(lim.Cancel, ShouldNotPanic)
		})

		Convey("Acquire can be cancelled", func() {
			lim := multilimiter.NewConcLimiter(DEFAULT_CONCURRENCY)
			timeout := time.Millisecond * 100

			err := lim.Acquire(Context(timeout))
			So(err, ShouldBeNil)

			lim.Cancel()

			err = lim.Acquire(Context(timeout))
			So(err, ShouldEqual, multilimiter.LimiterStopped)
		})

		Convey("Acquire adheres to a timeout", func() {
			timeout := 20 * time.Millisecond
			acquireDelay := timeout * 2 // ensure we should get a timeout

			lim := multilimiter.NewConcLimiter(1)

			done := make(chan struct{})

			var err1, err2 error
			go func() {
				err1 = lim.Acquire(Context(timeout))

				go func() {
					err2 = lim.Acquire(Context(timeout))
				}()

				time.Sleep(acquireDelay)

				lim.Release()
				close(done)
			}()

			<-done

			So(err1, ShouldBeNil)
			So(err2, ShouldEqual, multilimiter.DeadlineExceeded)
		})

		Convey("Acquire allows a 0 timeout", func() {
			lim := multilimiter.NewConcLimiter(DEFAULT_CONCURRENCY)
			err := lim.Acquire(Context(time.Millisecond * 0))
			So(err, ShouldBeNil)
		})
	})

}
