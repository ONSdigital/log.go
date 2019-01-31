package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSeverity(t *testing.T) {
	Convey("severity can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Severity, ShouldBeNil)

		FATAL.attach(event)
		So(event.Severity, ShouldNotBeNil)
		So(*event.Severity, ShouldEqual, FATAL)
	})

	Convey("severity values are of type severity", t, func() {
		So(FATAL, ShouldHaveSameTypeAs, severity(-1))
		So(ERROR, ShouldHaveSameTypeAs, severity(-1))
		So(WARN, ShouldHaveSameTypeAs, severity(-1))
		So(INFO, ShouldHaveSameTypeAs, severity(-1))
	})

	Convey("severity values match logging spec", t, func() {
		So(FATAL, ShouldEqual, 0)
		So(ERROR, ShouldEqual, 1)
		So(WARN, ShouldEqual, 2)
		So(INFO, ShouldEqual, 3)
	})
}
