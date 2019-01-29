package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestError(t *testing.T) {
	Convey("*eventError can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		err := eventError{}
		err.attach(event)

		So(event.Error, ShouldResemble, &err)
	})

	// TODO
	// stack trace
	// error type (struct vs other)
	// stringified error
}
