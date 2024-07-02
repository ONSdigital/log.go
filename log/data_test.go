package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestData(t *testing.T) {
	Convey("*Data can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		data := Data{}
		data.attach(event)

		So(event.Data, ShouldResemble, &data)
	})
}
