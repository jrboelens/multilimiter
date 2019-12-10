package multilimiter_test

import (
	"testing"
	"time"

	"github.com/jrboelens/multilimiter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCancelableSpec(t *testing.T) {

	Convey("Cancelable tests ", t, func() {

		Convey("Cancel is idempotent", func() {
			canceler := multilimiter.NewCanceler()

			So(canceler.IsCanceled(), ShouldEqual, false)

			alreadyCanceled := canceler.Cancel()

			// ensure we're now stopped and that we've flipped to a stopped state
			So(alreadyCanceled, ShouldEqual, false)
			So(canceler.IsCanceled(), ShouldEqual, true)

			alreadyCanceled = canceler.Cancel()

			// ensure we were already stopped
			So(alreadyCanceled, ShouldEqual, true)
			So(canceler.IsCanceled(), ShouldEqual, true)
		})

		Convey("The Done() channel is closed after Cancel is called", func() {
			canceler := multilimiter.NewCanceler()
			So(canceler.IsCanceled(), ShouldEqual, false)

			canceler.Cancel()

			select {
			case <-canceler.Done():
			case <-time.After(50 * time.Millisecond):
				// if we found ourselves here, we've failed
				So(false, ShouldEqual, true)
			}
		})
	})

}
