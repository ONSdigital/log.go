package log

import (
	"context"
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	request "github.com/ONSdigital/dp-net/request"

	. "github.com/smartystreets/goconvey/convey"
)

type writer struct {
	write func(b []byte) (n int, err error)
}

func (w writer) Write(b []byte) (n int, err error) {
	if w.write != nil {
		return w.write(b)
	}

	return 0, nil
}

func TestLog(t *testing.T) {
	Convey("Package defaults are right", t, func() {
		Convey("Namespace defaults to os.Args[0]", func() {
			So(Namespace, ShouldEqual, os.Args[0])
		})

		Convey("destination defaults to os.Stdout", func() {
			// This test is commented out because when running in test mode, it appears
			// that os.Stdout gets replaced (after destination is initialised), so they're
			// never equal.
			//
			// I'm leaving it in to show the intent, even if it can't be verified by the test

			//So(destination, ShouldEqual, os.Stdout)
		})

		Convey("fallbackDestination defaults to os.Stderr", func() {
			// This test is commented out because when running in test mode, it appears
			// that os.Stderr gets replaced (after fallbackDestination is initialised), so they're
			// never equal.
			//
			// I'm leaving it in to show the intent, even if it can't be verified by the test

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
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {
				wasCalled = true
			}}
			Event(nil, "", INFO)
			So(wasCalled, ShouldBeTrue)
		})

		Convey("Info calls eventFuncInst.f", func() {
			var wasCalled bool
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {
				wasCalled = true
			}}
			Info(nil, "", INFO)
			So(wasCalled, ShouldBeTrue)
		})

		Convey("Warn calls eventFuncInst.f", func() {
			var wasCalled bool
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {
				wasCalled = true
			}}
			Warn(nil, "", WARN)
			So(wasCalled, ShouldBeTrue)
		})

		Convey("ErrorDetails calls eventFuncInst.f", func() {
			var wasCalled bool
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {
				wasCalled = true
			}}
			ErrorDetails(nil, "", errors.New("error"), ERROR)
			So(wasCalled, ShouldBeTrue)
		})

		Convey("Fatal calls eventFuncInst.f", func() {
			var wasCalled bool
			eventFuncInst = &eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {
				wasCalled = true
			}}
			Fatal(nil, "", errors.New("fatal error"), FATAL)
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

	Convey("eventWithOptionsCheck panics if the same option is passed multiple times", t, func() {
		So(func() {
			eventWithOptionsCheck(nil, "event", INFO, Data{}, Data{})
		}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/v2/log.Data")
		So(func() {
			eventWithOptionsCheck(nil, "event", FATAL, INFO)
		}, ShouldPanicWith, "can't pass severity as a parameter")

		Convey("The first duplicate argument causes the panic", func() {
			So(func() {
				eventWithOptionsCheck(nil, "event", FATAL, Data{}, &EventHTTP{}, Data{})
			}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/v2/log.Data")
		})
	})

	Convey("eventWithOptionsCheck calls eventWithoutOptionsCheckFunc.f for valid arguments", t, func() {
		old := eventWithoutOptionsCheckFunc.f
		defer func() {
			eventWithoutOptionsCheckFunc.f = old
		}()

		var c context.Context
		var e string
		var o []option
		var called bool

		eventWithoutOptionsCheckFunc.f = func(ctx context.Context, event string, severity severity, opts ...option) {
			called = true
			c = ctx
			e = event
			o = opts
		}

		ctx := context.Background()
		So(called, ShouldBeFalse)

		eventWithOptionsCheck(ctx, "test event", INFO, Data{"value": 1})
		So(called, ShouldBeTrue)
		So(c, ShouldEqual, ctx)
		So(e, ShouldEqual, "test event")
		So(o, ShouldHaveLength, 1)
		So(o[0], ShouldHaveSameTypeAs, Data{})
		So(o[0], ShouldResemble, Data{"value": 1})
	})

	Convey("createEvent creates a new event", t, func() {

		Convey("createEvent should set the namespace", func() {
			evt := createEvent(nil, "event")
			So(evt.Namespace, ShouldEqual, Namespace)
		})

		Convey("createEvent should set the timestamp", func() {
			evt := createEvent(nil, "event")
			So(evt.CreatedAt.Unix(), ShouldBeGreaterThan, 0)

			now := time.Now().UTC()
			diff := now.Sub(evt.CreatedAt)
			// if this starts failing, and the code hasn't changed, check that
			// the two lines above actually take less than 100 milliseconds
			// (this should generally be true)
			//
			// all we really care about (for the test) is that the timestamp
			// has been set to a relatively recent value (and isn't hardcoded)
			So(diff, ShouldBeLessThan, time.Millisecond*100)
		})

		Convey("createEvent should set the event", func() {
			evt := createEvent(nil, "event")
			So(evt.Event, ShouldEqual, "event")

			evt = createEvent(nil, "test")
			So(evt.Event, ShouldEqual, "test")
		})

		Convey("createEvent sets the TraceID field to the request ID in the context", func() {
			ctx := request.WithRequestId(context.Background(), "trace ID")
			evt := createEvent(ctx, "event")
			So(evt.TraceID, ShouldEqual, "trace ID")

			ctx = request.WithRequestId(context.Background(), "another ID")
			evt = createEvent(ctx, "event")
			So(evt.TraceID, ShouldEqual, "another ID")
		})

		Convey("createEvent attaches options to the parent event", func() {
			evt := createEvent(nil, "event")
			So(evt.Auth, ShouldBeNil)

			e := Auth(USER, "identity")
			evt = createEvent(nil, "event", e)
			So(evt.Auth, ShouldEqual, e)
		})

	})

	Convey("print writes to stdout, or stderr on failure", t, func() {
		oldDestination := destination
		oldFallbackDestination := fallbackDestination

		defer func() {
			destination = oldDestination
			fallbackDestination = oldFallbackDestination
		}()

		var destCalled, fallbackDestCalled, destIsError bool

		destination = &writer{func(b []byte) (n int, err error) {
			destCalled = true
			if destIsError {
				return 0, errors.New("error")
			}
			return len(b), nil
		}}
		fallbackDestination = &writer{func(b []byte) (n int, err error) {
			fallbackDestCalled = true
			return len(b), nil
		}}

		Convey("empty slice does nothing", func() {
			So(destCalled, ShouldBeFalse)
			So(fallbackDestCalled, ShouldBeFalse)

			print([]byte{})

			So(destCalled, ShouldBeFalse)
			So(fallbackDestCalled, ShouldBeFalse)
		})

		Convey("non-empty slice writes to stdout", func() {
			So(destCalled, ShouldBeFalse)
			So(fallbackDestCalled, ShouldBeFalse)

			print([]byte("test"))

			So(destCalled, ShouldBeTrue)
			So(fallbackDestCalled, ShouldBeFalse)
		})

		Convey("non-empty slice writes to stderr if stdout errors", func() {
			So(destCalled, ShouldBeFalse)
			So(fallbackDestCalled, ShouldBeFalse)

			destIsError = true
			defer func() {
				destIsError = false
			}()

			print([]byte("test"))

			So(destCalled, ShouldBeTrue)
			So(fallbackDestCalled, ShouldBeTrue)
		})

		Convey("panic and exit if stdout and stderr are both closed", func() {
			// it's not possible to test this scenario
			// see the comment in log.go
		})
	})

	Convey("handleStyleError handles errors when serialising a log event", t, func() {

		Convey("return same bytes if error is nil", func() {
			b := []byte("test")
			b2 := handleStyleError(nil, EventData{}, eventFunc{nil}, b, nil)
			So(b, ShouldResemble, b2)
		})

		Convey("call eventFunc.f if error is not nil", func() {
			var called bool
			var calledCtx context.Context
			var calledEvent string
			var calledOpts []option
			f := func(ctx context.Context, event string, severity severity, opts ...option) {
				called = true
				calledCtx = ctx
				calledEvent = event
				calledOpts = opts
			}

			b := []byte("test")

			ctx := context.Background()

			So(called, ShouldBeFalse)
			isTestMode = false
			b2 := handleStyleError(ctx, EventData{}, eventFunc{f}, b, errors.New("test"))
			isTestMode = true
			So(called, ShouldBeTrue)
			So(b2, ShouldBeEmpty)

			So(calledCtx, ShouldEqual, ctx)
			So(calledEvent, ShouldEqual, "error marshalling event data")
			So(calledOpts, ShouldHaveLength, 2)

			So(calledOpts[0], ShouldHaveSameTypeAs, &EventError{})
			ee := calledOpts[0].(*EventError)
			So(ee.Message, ShouldEqual, "test")
			// ee.Data is an *errors.errorString, because it was made with errors.New()
			So(ee.Data, ShouldHaveSameTypeAs, errors.New("test"))
			So(ee.Data.(error).Error(), ShouldEqual, "test")

			So(calledOpts[1], ShouldHaveSameTypeAs, Data{})
			d := calledOpts[1].(Data)
			So(d, ShouldContainKey, "event_data")
			So(d["event_data"], ShouldEqual, "{CreatedAt:0001-01-01 00:00:00 +0000 UTC Namespace: Event: TraceID: SpanID: Severity:<nil> HTTP:<nil> Auth:<nil> Data:<nil> Error:<nil>}")
		})

		Convey("panic if running in test mode", func() {
			So(func() {
				handleStyleError(nil, EventData{}, eventFunc{func(ctx context.Context, event string, severity severity, opts ...option) {}}, []byte("test"), errors.New("test"))
			}, ShouldPanicWith, "error marshalling event data: {CreatedAt:0001-01-01 00:00:00 +0000 UTC Namespace: Event: TraceID: SpanID: Severity:<nil> HTTP:<nil> Auth:<nil> Data:<nil> Error:<nil>}")
		})

	})

	Convey("styleForMachine outputs JSON Lines format", t, func() {
		b := styleForMachine(nil, EventData{}, eventFunc{nil})
		// note: it's possible that json.Marshal won't always output the fields
		// 		 in the same order - can't think of a great solution atm
		So(string(b), ShouldResemble, "{\"created_at\":\"0001-01-01T00:00:00Z\",\"namespace\":\"\",\"event\":\"\"}")
	})

	Convey("styleForHuman outputs pretty printed JSON format", t, func() {
		b := styleForHuman(nil, EventData{}, eventFunc{nil})
		// note: it's possible that json.Marshal won't always output the fields
		// 		 in the same order - can't think of a great solution atm
		So(string(b), ShouldResemble, "{\n  \"created_at\": \"0001-01-01T00:00:00Z\",\n  \"event\": \"\",\n  \"namespace\": \"\"\n}")
	})

	Convey("eventWithoutOptionsCheck calls print with the output of the selected styler", t, func() {
		oldDestination := destination

		defer func() {
			destination = oldDestination
		}()

		styler = &struct {
			f func(context.Context, EventData, eventFunc) []byte
		}{func(context.Context, EventData, eventFunc) []byte {
			return []byte("styled output")
		}}

		var bytesWritten []byte
		destination = &writer{func(b []byte) (n int, err error) {
			bytesWritten = b
			return len(b), nil
		}}

		eventWithoutOptionsCheck(nil, "test", INFO)

		So(string(bytesWritten), ShouldResemble, "styled output\n")
	})
}
