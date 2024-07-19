package log_test

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/ONSdigital/log.go/v3/log"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/xerrors"

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
			fmtdErrors := log.FormatAsErrors(berr)
			So(fmtdErrors, ShouldHaveLength, 1)
			So(fmtdErrors[0].Message, ShouldEqual, berr.Error())

			fmtdErrors = log.FormatAsErrors(bwerr)
			So(fmtdErrors, ShouldHaveLength, 2)
			So(fmtdErrors[0].Message, ShouldEqual, bwerr.Error())
			So(fmtdErrors[1].Message, ShouldEqual, berr.Error())

			fmtdErrors = log.FormatAsErrors(bw2err)
			So(fmtdErrors, ShouldHaveLength, 3)
			So(fmtdErrors[0].Message, ShouldEqual, bw2err.Error())
			So(fmtdErrors[1].Message, ShouldEqual, bwerr.Error())
			So(fmtdErrors[2].Message, ShouldEqual, berr.Error())
		})

		Convey("pkg errors generate appropriate stacktrace", func() {
			fmtdErrors := log.FormatAsErrors(perr)
			So(fmtdErrors, ShouldHaveLength, 1)
			So(fmtdErrors[0].Message, ShouldEqual, perr.Error())
			So(len(fmtdErrors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			fmtdErrors = log.FormatAsErrors(pwerr)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(fmtdErrors, ShouldHaveLength, 3)
			So(fmtdErrors[0].Message, ShouldEqual, pwerr.Error())
			So(len(fmtdErrors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+10)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[1].Message, ShouldEqual, pwerr.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 0)
			So(fmtdErrors[2].Message, ShouldEqual, perr.Error())
			So(len(fmtdErrors[2].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[2].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(fmtdErrors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			fmtdErrors = log.FormatAsErrors(pw2err)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(fmtdErrors, ShouldHaveLength, 5)

			So(fmtdErrors[0].Message, ShouldEqual, pw2err.Error())
			So(len(fmtdErrors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+11)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[1].Message, ShouldEqual, pw2err.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 0)

			So(fmtdErrors[2].Message, ShouldEqual, pwerr.Error())
			So(len(fmtdErrors[2].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[2].StackTrace[0].Line, ShouldEqual, datumline+10)
			So(fmtdErrors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[3].Message, ShouldEqual, pwerr.Error())
			So(fmtdErrors[3].StackTrace, ShouldHaveLength, 0)

			So(fmtdErrors[4].Message, ShouldEqual, perr.Error())
			So(len(fmtdErrors[4].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[4].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[4].StackTrace[0].Line, ShouldEqual, datumline+9)
			So(fmtdErrors[4].StackTrace[0].Function, ShouldEqual, datumFunc)
		})

		Convey("golang xerrors generate appropriate stacktrace", func() {
			fmtdErrors := log.FormatAsErrors(xerr)
			So(fmtdErrors, ShouldHaveLength, 1)
			So(fmtdErrors[0].Message, ShouldEqual, xerr.Error())
			// xerrror stack traces are only 1 level deep
			So(fmtdErrors[0].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			fmtdErrors = log.FormatAsErrors(xwerr)
			So(fmtdErrors, ShouldHaveLength, 2)
			So(fmtdErrors[0].Message, ShouldEqual, xwerr.Error())
			So(fmtdErrors[0].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+15)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[1].Message, ShouldEqual, xerr.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[1].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(fmtdErrors[1].StackTrace[0].Function, ShouldEqual, datumFunc)

			fmtdErrors = log.FormatAsErrors(xw2err)
			So(fmtdErrors, ShouldHaveLength, 3)
			So(fmtdErrors[0].Message, ShouldEqual, xw2err.Error())
			So(fmtdErrors[0].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+16)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[1].Message, ShouldEqual, xwerr.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[1].StackTrace[0].Line, ShouldEqual, datumline+15)
			So(fmtdErrors[1].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[2].Message, ShouldEqual, xerr.Error())
			So(fmtdErrors[2].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[2].StackTrace[0].Line, ShouldEqual, datumline+14)
			So(fmtdErrors[2].StackTrace[0].Function, ShouldEqual, datumFunc)
		})

		Convey("mixed errors generate appropriate stacktrace", func() {
			fmtdErrors := log.FormatAsErrors(pb)
			So(fmtdErrors, ShouldHaveLength, 3)
			So(fmtdErrors[0].Message, ShouldEqual, pb.Error())
			So(len(fmtdErrors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+19)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[1].Message, ShouldEqual, pb.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 0)
			So(fmtdErrors[2].Message, ShouldEqual, berr.Error())
			So(fmtdErrors[2].StackTrace, ShouldHaveLength, 0)

			fmtdErrors = log.FormatAsErrors(xb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(fmtdErrors, ShouldHaveLength, 2)
			So(fmtdErrors[0].Message, ShouldEqual, xb.Error())
			So(fmtdErrors[0].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+20)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)
			So(fmtdErrors[1].Message, ShouldEqual, berr.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 0)

			fmtdErrors = log.FormatAsErrors(xpb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(fmtdErrors, ShouldHaveLength, 4)

			So(fmtdErrors[0].Message, ShouldEqual, xpb.Error())
			So(fmtdErrors[0].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+21)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[1].Message, ShouldEqual, pb.Error())
			So(len(fmtdErrors[1].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[1].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[1].StackTrace[0].Line, ShouldEqual, datumline+19)
			So(fmtdErrors[1].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[2].Message, ShouldEqual, pb.Error())
			So(fmtdErrors[2].StackTrace, ShouldHaveLength, 0)

			So(fmtdErrors[3].Message, ShouldEqual, berr.Error())
			So(fmtdErrors[3].StackTrace, ShouldHaveLength, 0)

			fmtdErrors = log.FormatAsErrors(pxb)
			// pkg wraps errors weirdly so there is double wrapping with stack-traced and non-stack-traced errors
			So(fmtdErrors, ShouldHaveLength, 4)

			So(fmtdErrors[0].Message, ShouldEqual, pxb.Error())
			So(len(fmtdErrors[0].StackTrace), ShouldBeGreaterThan, 1)
			So(fmtdErrors[0].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[0].StackTrace[0].Line, ShouldEqual, datumline+22)
			So(fmtdErrors[0].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[1].Message, ShouldEqual, pxb.Error())
			So(fmtdErrors[1].StackTrace, ShouldHaveLength, 0)

			So(fmtdErrors[2].Message, ShouldEqual, xb.Error())
			So(fmtdErrors[2].StackTrace, ShouldHaveLength, 1)
			So(fmtdErrors[2].StackTrace[0].File, ShouldEqual, datumFile)
			So(fmtdErrors[2].StackTrace[0].Line, ShouldEqual, datumline+20)
			So(fmtdErrors[2].StackTrace[0].Function, ShouldEqual, datumFunc)

			So(fmtdErrors[3].Message, ShouldEqual, berr.Error())
			So(fmtdErrors[3].StackTrace, ShouldHaveLength, 0)
		})
	})
}
