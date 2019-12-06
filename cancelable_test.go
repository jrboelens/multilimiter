package multilimiter_test

import (
	"testing"

	"github.com/jrboelens/multilimiter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCancelableSpec(t *testing.T) {

	Convey("Cancelable tests ", t, func() {

		Convey("Cancel is idempotent", func() {
			canceler := multilimiter.Canceler{}

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

		Convey("a function can be called as part of cancellation if the state flips to canceled", func() {
			canceler := multilimiter.Canceler{}
			So(canceler.IsCanceled(), ShouldEqual, false)

			changedCount := 0
			fn := func() { changedCount++ }

			canceler.Cancel(fn)
			So(changedCount, ShouldEqual, 1)

			// we don't see an increment here because the canceled state didn't change
			canceler.Cancel(fn)
			So(changedCount, ShouldEqual, 1)
		})
	})

}
