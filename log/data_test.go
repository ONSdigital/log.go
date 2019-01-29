package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestData(t *testing.T) {
	/*
		TODO

		Somehow test that Data{} is a map[string]interface{}

		More specifically, it must always marshall to an object
		type in JSON - the actual underlying type doesn't matter
	*/

	Convey("*Data can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		data := Data{}
		data.attach(event)

		So(event.Data, ShouldResemble, &data)
	})
}
