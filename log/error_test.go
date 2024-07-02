package log_test

import (
	"errors"
	"fmt"
	"github.com/ONSdigital/log.go/v3/log"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/xerrors"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatAsErrors(t *testing.T) {
	Convey("with a number of predefined errors", t, func() {

		// line number of here used as datum for generated stack traces
		pc, datumFile, datumline, _ := runtime.Caller(0)
		datumFunc := runtime.FuncForPC(pc).Name()

		// basic go errors
		berr := errors.New("basic error")
		bwerr := fmt.Errorf("basic wrapped error [%w]", berr)
		bw2err := fmt.Errorf("double wrapped error [%w]", bwerr)

		// pkg/errors
		perr := pkgerrors.New("pkg error")
		pwerr := pkgerrors.Wrap(perr, "pkg wrapped error")
		pw2err := pkgerrors.Wrap(pwerr, "double pkg wrapped error")

		// golang.org/x/xerror
		xerr := xerrors.New("x error")
		xwerr := xerrors.Errorf("x wrapped %w", xerr)
		xw2err := xerrors.Errorf("x double wrapped %w", xwerr)

		// Mixed errors
		pb := pkgerrors.Wrap(berr, "pkg wrapped")
		xb := xerrors.Errorf("x wrapped %w", berr)
		xpb := xerrors.Errorf("x wrapped %w", pb)
		pxb := pkgerrors.Wrap(xb, "pkg wrapped")

		Convey("basic errors generate no stacktrace", func() {
			errors := log.FormatAsErrors(berr)
			So(errors, ShouldHaveLength, 1)
			So(errors[0].Message, ShouldEqual, berr.Error())

			errors = log.FormatAsErrors(bwerr)
			So(errors, ShouldHaveLength, 2)
			So(errors[0].Message, ShouldEqual, bwerr.Error())
			So(errors[1].Message, ShouldEqual, berr.Error())

			errors = log.FormatAsErrors(bw2err)
			So(errors, ShouldHaveLength, 3)
			So(errors[0].Message, ShouldEqual, bw2err.Error())
			So(errors[1].Message, ShouldEqual, bwerr.Error())
			So(errors[2].Message, ShouldEqual, berr.Error())
		})

		Convey("pkg errors generate appropriate stacktrace", func() {
			errors := log.FormatAsErrors(perr)
			So(errors, ShouldHaveLength, 1)
			So(errors[0].Message, ShouldEqual, perr.Error())
			So(len(errors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			errors = log.FormatAsErrors(pwerr)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(errors, ShouldHaveLength, 3)
			So(errors[0].Message, ShouldEqual, pwerr.Error())
			So(len(errors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+10)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[1].Message, ShouldEqual, pwerr.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 0)
			So(errors[2].Message, ShouldEqual, perr.Error())
			So(len(errors[2].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[2].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(errors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			errors = log.FormatAsErrors(pw2err)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(errors, ShouldHaveLength, 5)

			So(errors[0].Message, ShouldEqual, pw2err.Error())
			So(len(errors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+11)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[1].Message, ShouldEqual, pw2err.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 0)

			So(errors[2].Message, ShouldEqual, pwerr.Error())
			So(len(errors[2].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[2].StackTrace[0].Line, ShouldEqual, datumline+10)
			So(errors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[3].Message, ShouldEqual, pwerr.Error())
			So(errors[3].StackTrace, ShouldHaveLength, 0)

			So(errors[4].Message, ShouldEqual, perr.Error())
			So(len(errors[4].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[4].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[4].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(errors[4].StackTrace[0].Function, ShouldEqual, datumFunc)
		})

		Convey("golang xerrors generate appropriate stacktrace", func() {
			errors := log.FormatAsErrors(xerr)
			So(errors, ShouldHaveLength, 1)
			So(errors[0].Message, ShouldEqual, xerr.Error())
			// xerrror stack traces are only 1 level deep
			So(errors[0].StackTrace, ShouldHaveLength, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			errors = log.FormatAsErrors(xwerr)
			So(errors, ShouldHaveLength, 2)
			So(errors[0].Message, ShouldEqual, xwerr.Error())
			So(errors[0].StackTrace, ShouldHaveLength, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+15)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[1].Message, ShouldEqual, xerr.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 1)
			So(errors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[1].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(errors[1].StackTrace[0].Function, ShouldEqual, datumFunc)

			errors = log.FormatAsErrors(xw2err)
			So(errors, ShouldHaveLength, 3)
			So(errors[0].Message, ShouldEqual, xw2err.Error())
			So(errors[0].StackTrace, ShouldHaveLength, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+16)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[1].Message, ShouldEqual, xwerr.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 1)
			So(errors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[1].StackTrace[0].Line, ShouldEqual, datumline+15)
			So(errors[1].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[2].Message, ShouldEqual, xerr.Error())
			So(errors[2].StackTrace, ShouldHaveLength, 1)
			So(errors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[2].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(errors[2].StackTrace[0].Function, ShouldEqual, datumFunc)
		})

		Convey("mixed errors generate appropriate stacktrace", func() {
			errors := log.FormatAsErrors(pb)
			So(errors, ShouldHaveLength, 3)
			So(errors[0].Message, ShouldEqual, pb.Error())
			So(len(errors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+19)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[1].Message, ShouldEqual, pb.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 0)
			So(errors[2].Message, ShouldEqual, berr.Error())
			So(errors[2].StackTrace, ShouldHaveLength, 0)

			errors = log.FormatAsErrors(xb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(errors, ShouldHaveLength, 2)
			So(errors[0].Message, ShouldEqual, xb.Error())
			So(errors[0].StackTrace, ShouldHaveLength, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+20)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(errors[1].Message, ShouldEqual, berr.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 0)

			errors = log.FormatAsErrors(xpb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(errors, ShouldHaveLength, 4)

			So(errors[0].Message, ShouldEqual, xpb.Error())
			So(errors[0].StackTrace, ShouldHaveLength, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+21)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[1].Message, ShouldEqual, pb.Error())
			So(len(errors[1].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[1].StackTrace[0].Line, ShouldEqual, datumline+19)
			So(errors[1].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[2].Message, ShouldEqual, pb.Error())
			So(errors[2].StackTrace, ShouldHaveLength, 0)

			So(errors[3].Message, ShouldEqual, berr.Error())
			So(errors[3].StackTrace, ShouldHaveLength, 0)

			errors = log.FormatAsErrors(pxb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(errors, ShouldHaveLength, 4)

			So(errors[0].Message, ShouldEqual, pxb.Error())
			So(len(errors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(errors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[0].StackTrace[0].Line, ShouldEqual, datumline+22)
			So(errors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[1].Message, ShouldEqual, pxb.Error())
			So(errors[1].StackTrace, ShouldHaveLength, 0)

			So(errors[2].Message, ShouldEqual, xb.Error())
			So(errors[2].StackTrace, ShouldHaveLength, 1)
			So(errors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(errors[2].StackTrace[0].Line, ShouldEqual, datumline+20)
			So(errors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(errors[3].Message, ShouldEqual, berr.Error())
			So(errors[3].StackTrace, ShouldHaveLength, 0)

		})
	})
}
