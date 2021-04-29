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
		err := FormatErrors([]error{errors.New("test")})
		So(err, ShouldHaveSameTypeAs, &EventErrors{})
		So(err, ShouldImplement, (*option)(nil))

		Convey("*EventError has the correct fields", func() {
			myErr := []error{errors.New("test")}
			myData := Data{"value": []error{errors.New("test")}}
			ee := FormatErrors(myErr).(*EventErrors)
			So((*ee)[0].Message, ShouldEqual, "test")
			So((*ee)[0].Data, ShouldResemble, myData)
			So((*ee)[0].StackTrace, ShouldHaveLength, 10)
		})
	})

	Convey("*EventError can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		err := EventErrors{}
		err.attach(event)

		So(event.Errors, ShouldResemble, &err)
	})

	Convey("Message function sets *EventError.Message to error.Message()", t, func() {
		err := errors.New("test error")
		errEventData := FormatErrors([]error{err}).(*EventErrors)
		So((*errEventData)[0].Message, ShouldEqual, "test error")

		err = customError{"goodbye"}
		errEventData = FormatErrors([]error{err}).(*EventErrors)
		So((*errEventData)[0].Message, ShouldEqual, "hi there!")
	})

	Convey("Message function sets *EventError.Data to error", t, func() {
		Convey("A value of kind 'Struct' is embedded directly", func() {
			err := customError{}
			// data := Data{}
			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Data, ShouldHaveSameTypeAs, err)
		})
		Convey("A value of kind 'Ptr->Struct' is embedded directly", func() {
			err := customError{}
			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Data, ShouldEqual, err)
		})
		Convey("A value of other kinds (e.g. 'Int') is wrapped in Data{value:<err>}", func() {
			err := customIntError(0)
			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Data, ShouldHaveSameTypeAs, Data{})
			So((*errEventData)[0].Data.(Data)["value"], ShouldHaveSameTypeAs, customIntError(0))
		})
	})

	Convey("Message function generates a stack trace", t, func() {
		err := errors.New("new error")
		// WARNING if this line moves, update `So(origin.Line, ...)` below
		errEventData := FormatErrors([]error{err}).(*EventErrors)
		So((*errEventData)[0].StackTrace, ShouldHaveLength, 10)
		origin := (*errEventData)[0].StackTrace[0]
		So(origin.File, ShouldEndWith, "log.go/log/error_test.go")
		// If this test fails, check the `errEventData := Error(err).(*EventError)` line is still line 81!
		So(origin.Line, ShouldEqual, 83)
		So(origin.Function, ShouldEqual, "github.com/ONSdigital/log.go/v2/log.TestError.func5")
	})
}
