package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClassification(t *testing.T) {
	Convey("Classification can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Classification, ShouldBeEmpty)

		opt := Classification(ProtectiveMonitoring)
		opt.attach(event)

		So(string(event.Classification), ShouldEqual, "PROTECTIVE_MONITORING")
	})
}
