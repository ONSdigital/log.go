package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClassification(t *testing.T) {
	Convey("Classification can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Classification, ShouldBeEmpty)

		opt := Classification("PROTECTIVE_MONITORING")
		opt.attach(event)

		So(event.Classification, ShouldEqual, "PROTECTIVE_MONITORING")
	})
}
