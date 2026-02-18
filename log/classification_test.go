package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClassification(t *testing.T) {
	Convey("Classification can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Classification, ShouldBeEmpty)

		classification := Classification("PROTECTIVE_MONITORING")
		classification.attach(event)

		So(event.Classification, ShouldResemble, string(classification))
	})
}
