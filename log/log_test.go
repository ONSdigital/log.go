package log

import (
	"context"
	"flag"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLog(t *testing.T) {
	Convey("Package defaults are right", t, func() {
		Convey("Namespace defaults to os.Args[0]", func() {
			So(Namespace, ShouldEqual, os.Args[0])
		})

		Convey("destination defaults to os.Stdout", func() {
			//So(destination, ShouldEqual, os.Stdout)
		})

		Convey("fallbackDestination defaults to os.Stderr", func() {
			//So(destination, ShouldEqual, os.Stderr)
		})

		Convey("Package detects test mode", func() {
			Convey("Test mode off by default", func() {
				oldCommandLine := flag.CommandLine
				defer func() {
					flag.CommandLine = oldCommandLine
				}()
				flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
				f := initEvent()
				So(f, ShouldEqual, eventWithoutOptionsCheckFunc)
				So(isTestMode, ShouldBeFalse)
			})

			Convey("Test mode on if test.v flag exists", func() {
				oldCommandLine := flag.CommandLine
				defer func() {
					flag.CommandLine = oldCommandLine
				}()
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				flag.CommandLine.Bool("test.v", true, "")
				f := initEvent()
				So(f, ShouldEqual, eventWithOptionsCheckFunc)
				So(isTestMode, ShouldBeTrue)
			})
		})

		Convey("Event calls eventFuncInst.f", func() {
			var wasCalled bool
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, opts ...option) {
				wasCalled = true
			}}
			Event(nil, "")
			So(wasCalled, ShouldBeTrue)
		})

		Convey("styler function is set correctly", func() {
			Convey("styler is set to styleForMachineFunc by default", func() {
				So(initStyler(), ShouldEqual, styleForMachineFunc)
			})
			Convey("styler is set to styleForHumanFunc if HUMAN_LOG environment variable is set", func() {
				oldValue := os.Getenv("HUMAN_LOG")
				os.Setenv("HUMAN_LOG", "1")
				So(initStyler(), ShouldEqual, styleForHumanFunc)
				os.Setenv("HUMAN_LOG", oldValue)
			})
		})
	})
}
