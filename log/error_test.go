package log

import (
	"errors"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type customIntError int

func (c customIntError) Error() string {
	return strconv.Itoa(int(c))
}

func TestFormatErrorsFunc(t *testing.T) {
	t.Parallel()

	// Keep this test function at top of test to help prevent the test failing. This is due to the test
	// assertion to check the line number of the stacktrace which is defined when calling FormatErrors func.
	Convey("Check error event data generates a stack trace", t, func() {
		err := errors.New("new error")

		// WARNING if this line moves, update `So(origin.Line, ...)` below
		errEventData := FormatErrors([]error{err}).(*EventErrors)
		So((*errEventData)[0].StackTrace, ShouldHaveLength, 10)

		origin := (*errEventData)[0].StackTrace[0]
		So(origin.File, ShouldEndWith, "log.go/log/error_test.go")

		// If this test fails, check the `errEventData := Error(err).(*EventErrors)` line is still line 26 of the this test file!
		So(origin.Line, ShouldEqual, 26)
		So(origin.Function, ShouldEqual, "github.com/ONSdigital/log.go/v2/log.TestFormatErrorsFunc.func1")
	})

	Convey("Check *EventErrors is returned and implements the option interface", t, func() {
		err := FormatErrors([]error{errors.New("test")})
		So(err, ShouldHaveSameTypeAs, &EventErrors{})
		So(err, ShouldImplement, (*option)(nil))

		Convey("Check *EventErrors contains the expected fields", func() {
			myErr := []error{errors.New("test")}

			errEventData := FormatErrors(myErr).(*EventErrors)
			So((*errEventData)[0].Message, ShouldEqual, myErr[0].Error())
			So((*errEventData)[0].Data, ShouldBeNil)
			So((*errEventData)[0].StackTrace, ShouldHaveLength, 10)
		})
	})

	Convey("Check *EventErrors can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Data, ShouldBeNil)

		err := EventErrors{}
		err.attach(event)

		So(event.Errors, ShouldResemble, &err)
	})

	Convey("Check event error Data is set to error", t, func() {
		Convey("For a value of kind 'Struct' is embedded directly via a custom error", func() {
			err := &CustomError{Message: "goodbye"}

			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Data, ShouldBeNil)
			So((*errEventData)[0].Message, ShouldEqual, "goodbye")
		})

		Convey("For a value of kind 'Ptr->Struct' is embedded directly", func() {
			err := &CustomError{
				Message: "new error",
				Data:    map[string]interface{}{"new_data_variable": "546"},
			}
			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Message, ShouldEqual, "new error")
			So((*errEventData)[0].Data, ShouldHaveSameTypeAs, err.Data)
			So((*errEventData)[0].Data, ShouldEqual, err.Data)
		})

		Convey("For a value of other kinds (e.g. 'Int') is wrapped in Data{value:<err>}", func() {
			err := customIntError(0)
			errEventData := FormatErrors([]error{err}).(*EventErrors)
			So((*errEventData)[0].Data, ShouldHaveSameTypeAs, Data{})
			So((*errEventData)[0].Data.(Data)["value"], ShouldEqual, err)
			So((*errEventData)[0].Message, ShouldEqual, "0")
		})
	})

	Convey("Check first two items in *EventErrors and contains the expected error event data", t, func() {
		err1 := errors.New("test error")
		err2 := &CustomError{
			Message: "hidden error",
			Data:    map[string]interface{}{"count": 12},
		}

		errEventData := FormatErrors([]error{err1, err2}).(*EventErrors)
		So(errEventData, ShouldHaveLength, 2)

		// First item in error event data
		So((*errEventData)[0].Data, ShouldBeNil)
		So((*errEventData)[0].Message, ShouldEqual, err1.Error())

		// Second item in error event data
		So((*errEventData)[1].Data, ShouldHaveSameTypeAs, make(map[string]interface{}))
		So((*errEventData)[1].Data, ShouldEqual, err2.Data)
		So((*errEventData)[1].Message, ShouldEqual, err2.Message)
	})

	Convey("If a nil value is passed into FormatErrors, check no error event data is returned", t, func() {
		errEventData := FormatErrors([]error{nil}).(*EventErrors)
		So(errEventData, ShouldHaveLength, 0)
	})
}
