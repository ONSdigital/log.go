package log

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type customError struct {
	CustomField string `json:"custom_field"`
}

func (c customError) Error() string {
	return "hi there!"
}

type customIntError int

func (c customIntError) Error() string {
	return "hello!"
}

func TestError(t *testing.T) {
	Convey("Error function returns a *EventError", t, func() {
		err := Error(errors.New("test"))
		So(err, ShouldHaveSameTypeAs, &EventError{})
		So(err, ShouldImplement, (*option)(nil))

		Convey("*EventError has the correct fields", func() {
			myErr := errors.New("test error")
			ee := Error(myErr).(*EventError)
			So(ee.Error, ShouldEqual, "test error")
			So(ee.Data, ShouldResemble, myErr)
			So(ee.StackTrace, ShouldHaveLength, 10)
		})
	})

	Convey("*EventError can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		err := EventError{}
		err.attach(event)

		So(event.Error, ShouldResemble, &err)
	})

	Convey("Error function sets *EventError.Error to error.Error()", t, func() {
		err := errors.New("test error")
		errEventData := Error(err).(*EventError)
		So(errEventData.Error, ShouldEqual, "test error")

		err = customError{"goodbye"}
		errEventData = Error(err).(*EventError)
		So(errEventData.Error, ShouldEqual, "hi there!")
	})

	Convey("Error function sets *EventError.Data to error", t, func() {
		Convey("A value of kind 'Struct' is embedded directly", func() {
			err := customError{}
			errEventData := Error(err).(*EventError)
			So(errEventData.Data, ShouldHaveSameTypeAs, err)
		})
		Convey("A value of kind 'Ptr->Struct' is embedded directly", func() {
			err := &customError{}
			errEventData := Error(err).(*EventError)
			So(errEventData.Data, ShouldEqual, err)
		})
		Convey("A value of other kinds (e.g. 'Int') is wrapped in Data{value:<err>}", func() {
			err := customIntError(0)
			errEventData := Error(err).(*EventError)
			So(errEventData.Data, ShouldHaveSameTypeAs, Data{})
			So(errEventData.Data.(Data)["value"], ShouldHaveSameTypeAs, customIntError(0))
		})
	})

	Convey("Error function generates a stack trace", t, func() {
		err := errors.New("new error")
		// WARNING if this line moves, update `So(origin.Line, ...)` below
		errEventData := Error(err).(*EventError)
		So(errEventData.StackTrace, ShouldHaveLength, 10)
		origin := errEventData.StackTrace[0]
		So(origin.File, ShouldEndWith, "log.go/log/error_test.go")
		// If this test fails, check the `errEventData := Error(err).(*EventError)` line is still line 81!
		So(origin.Line, ShouldEqual, 81)
		So(origin.Function, ShouldEqual, "github.com/ONSdigital/log.go/log.TestError.func5")
	})
}
