package log

import (
	"context"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGoLogger(t *testing.T) {
	var capCtx context.Context
	var capEvent string
	var capOpts []option
	var hasBeenCalled bool

	// Capture the existing function so we can reset it for other tests
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	// Replace existing function with a test
	Event = func(ctx context.Context, event string, opts ...option) {
		capCtx = ctx
		capEvent = event
		capOpts = opts
		hasBeenCalled = true
	}

	Convey("Log data from standard library logger is captured", t, func() {
		So(hasBeenCalled, ShouldBeFalse)
		So(capCtx, ShouldBeNil)
		So(capEvent, ShouldBeEmpty)
		So(capOpts, ShouldHaveLength, 0)

		log.Println("test")

		So(hasBeenCalled, ShouldBeTrue)
		So(capCtx, ShouldBeNil)
		So(capEvent, ShouldEqual, "third party logs")
		So(capOpts, ShouldHaveLength, 1)
		So(capOpts[0], ShouldHaveSameTypeAs, Data{})
		capData := capOpts[0].(Data)
		So(capData, ShouldContainKey, "raw")
		So(capData["raw"], ShouldHaveSameTypeAs, "example")
		So(capData["raw"], ShouldEqual, "test")
	})
}
