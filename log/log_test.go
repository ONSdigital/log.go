package log

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"

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

	Convey("eventWithOptionsCheck panics if the same option is passed multiple times", t, func() {
		So(func() {
			eventWithOptionsCheck(nil, "event", Data{}, Data{})
		}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/log.Data")
		So(func() {
			eventWithOptionsCheck(nil, "event", FATAL, INFO)
		}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/log.severity")

		Convey("The first duplicate argument causes the panic", func() {
			So(func() {
				eventWithOptionsCheck(nil, "event", FATAL, Data{}, INFO, Data{})
			}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/log.severity")
			So(func() {
				eventWithOptionsCheck(nil, "event", FATAL, Data{}, Data{}, INFO)
			}, ShouldPanicWith, "can't pass in the same parameter type multiple times: github.com/ONSdigital/log.go/log.Data")
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

		eventWithoutOptionsCheckFunc.f = func(ctx context.Context, event string, opts ...option) {
			called = true
			c = ctx
			e = event
			o = opts
		}

		ctx := context.Background()
		So(called, ShouldBeFalse)

		eventWithOptionsCheck(ctx, "test event", FATAL)

		So(called, ShouldBeTrue)
		So(c, ShouldEqual, ctx)
		So(e, ShouldEqual, "test event")
		So(o, ShouldHaveLength, 1)
		So(o[0], ShouldHaveSameTypeAs, INFO)
		So(o[0], ShouldEqual, FATAL)
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
			ctx := common.WithRequestId(context.Background(), "trace ID")
			evt := createEvent(ctx, "event")
			So(evt.TraceID, ShouldEqual, "trace ID")

			ctx = common.WithRequestId(context.Background(), "another ID")
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
			f := func(ctx context.Context, event string, opts ...option) {
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
			So(ee.Error, ShouldEqual, "test")
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
				handleStyleError(nil, EventData{}, eventFunc{func(ctx context.Context, event string, opts ...option) {}}, []byte("test"), errors.New("test"))
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

		eventWithoutOptionsCheck(nil, "test")

		So(string(bytesWritten), ShouldResemble, "styled output\n")
	})
}

// run with:
// go test -run=log_test.go -bench=Log -benchtime=100x

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
var requestIDRandom = rand.New(rand.NewSource(99)) // seed with constant to get same sequence out output for every benchmar run
var randMutex sync.Mutex

// NewRequestID generates a random string of requested length
func newRequestID(size int) string {
	b := make([]rune, size)
	randMutex.Lock()
	for i := range b {
		b[i] = letters[requestIDRandom.Intn(len(letters))]
	}
	randMutex.Unlock()
	return string(b)
}

func BenchmarkLog1(b *testing.B) {
	fmt.Println("Benchmarking: 'Log'")
	errToLog := errors.New("test error")
	message1 := "Benchmark test"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("req: %v\n", req)

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = false // use existing Event() code

	b.ReportAllocs()

	// test all event types
	for i := 0; i < b.N; i++ {
		Event(ctx,
			message1,
			INFO,
			Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4},
			Error(errToLog),
			HTTP(req, 0, 0, nil, nil),
			Auth(USER, "tester-1"))
	}
}

// run with:
// go test -run=log_test.go -bench=. -benchtime=1000000000x

// on 1st May 2020 gave results:
/*

Benchmarking: 'Log - o.attach'
goos: linux
goarch: amd64
pkg: github.com/ONSdigital/log.go/log
BenchmarkLog2-12    	Benchmarking: 'Log - o.attach'
1000000000	       142 ns/op
Benchmarking: 'Log - switch'
BenchmarkLog3-12    	Benchmarking: 'Log - switch'
1000000000	       141 ns/op
PASS
ok  	github.com/ONSdigital/log.go/log	282.637s

*/

func BenchmarkLog2(b *testing.B) {
	fmt.Println("Benchmarking: 'Log - o.attach'")
	err := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"

	var opts [4]option

	opts[0] = INFO
	opts[1] = Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4}
	opts[2] = Error(err)
	opts[3] = Data{"data_4": data4, "data_2": data2}

	e := EventData{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     message1,
	}

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	for i := 0; i < b.N; i++ {
		// loop around each log option and call its attach method, which takes care
		// of the association with the EventData struct
		for _, o := range opts {
			// Using rare pattern : `thing.attach(toObject)`
			// this handles both cases where:
			// the receiver can be called either `dataThing.attach(...)` or `ptrToDataThing.attach(...)
			o.attach(&e)
		}
	}
}

func BenchmarkLog3(b *testing.B) {
	fmt.Println("Benchmarking: 'Log - switch'")
	err := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"

	var opts [4]option

	opts[0] = INFO
	opts[1] = Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4}
	opts[2] = Error(err)
	opts[3] = Data{"data_4": data4, "data_2": data2}

	e := EventData{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     message1,
	}

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	for i := 0; i < b.N; i++ {
		// loop around each log option and attach each option
		// directly into EventData struct
		for _, o := range opts {
			// Doing typical pattern : `object.attach(thing)`
			switch v := o.(type) {
			case severity:
				e.Severity = &v
			case *severity: // added to match o.attach(e) code for completness (may never be used)
				e.Severity = v
			case Data:
				e.Data = &v
			case *Data: // added to match o.attach(e) code for completness (may never be used)
				e.Data = v
			case *EventHTTP:
				e.HTTP = v
			case *EventError:
				e.Error = v
			case *eventAuth:
				e.Auth = v
			default:
				fmt.Printf("option: %v, %v, %T", o, v, v)
				panic("unknown option")
			}
		}
	}
}

func BenchmarkLog4(b *testing.B) {
	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2 := req
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

		// 2nd event is 'similar in length' to one in createReverseProxy()
		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		// 3rd Event is like the second one in Middleware()
		Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))
	}
}

func BenchmarkLog5(b *testing.B) {
	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2 := req
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = false // use existing Event() code

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

		// 2nd event is 'similar in length' to one in createReverseProxy()
		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		// 3rd Event is like the second one in Middleware()
		Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))
	}
}

func BenchmarkLog6(b *testing.B) {
	fmt.Println("Benchmarking: 'Log'")
	errToLog := errors.New("test error")
	message1 := "Benchmark test"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("req: %v\n", req)

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()

	// test all event types
	for i := 0; i < b.N; i++ {
		Event(ctx,
			message1,
			INFO,
			Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4},
			Error(errToLog),
			HTTP(req, 0, 0, nil, nil),
			Auth(USER, "tester-1"))
	}
}

func BenchmarkLog7(b *testing.B) {
	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is revere proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2 := req
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		var statusCode int = 0

		port := 0
		if p := req.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}

		var duration *time.Duration

		// inline the the setting up of the "EventHTTP" to save doing the HTTP(...)
		// thing as this escapes to the heap, whereas doing the following stays within
		// the stack of this calling function.
		e := EventHTTP{
			StatusCode: &statusCode,
			Method:     req.Method,

			Scheme: req.URL.Scheme,
			Host:   req.URL.Hostname(),
			Port:   port,
			Path:   req.URL.Path,
			Query:  req.URL.RawQuery,

			StartedAt:             &start,
			EndedAt:               nil,
			Duration:              duration,
			ResponseContentLength: 0,
		}

		//Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))
		Event(ctx, "http request received", &e)

		port = 0
		if p := req2.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}

		e.Method = req2.Method
		e.Scheme = req2.URL.Scheme
		e.Host = req2.URL.Hostname()
		e.Port = port
		e.Path = req2.URL.Path
		e.Query = req2.URL.RawQuery
		e.StartedAt = nil

		// 2nd event is 'similar in length' to one in createReverseProxy()
		//		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Event(ctx, "proxying request", INFO, &e,
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		port = 0
		if p := req2.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}
		port = 20000

		d := end.Sub(start)

		e.Port = port
		e.StartedAt = &start
		e.EndedAt = &end
		e.Duration = &d
		e.ResponseContentLength = 4
		statusCode = 200

		e.Method = req2.Method
		e.Scheme = req2.URL.Scheme
		e.Host = req2.URL.Hostname()
		e.Port = port
		e.Path = req2.URL.Path
		e.Query = req2.URL.RawQuery

		// 3rd Event is like the second one in Middleware()
		//		Event(req.Context(), "http request completed", HTTP(req2, 200, 4, &start, &end))
		Event(ctx, "http request completed", &e)
	}
}

func TestLogNew1(t *testing.T) {
	// Test 3 events that look like what dp-frontend-router issues on the HAPPY HOT-PATH
	// Get the old events for the 3 and the new events for 3 and compare ...

	oldDestination := destination
	oldFallbackDestination := fallbackDestination

	defer func() {
		destination = oldDestination
		fallbackDestination = oldFallbackDestination
	}()

	fmt.Println("Testing: 'New Log 1'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2 := req
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	//////////////////////
	// Capture old events

	isMinimalAllocations = false // use existing Event() code

	// 1st Event is like the first one in Middleware()
	// Capture the output of the call to Event()
	var bytesWritten []byte
	destination = &writer{func(b []byte) (n int, err error) {
		bytesWritten = b
		return len(b), nil
	}}
	Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

	// Converting what has been captured in bytesWritten with string()
	// puts : !F(MISSING)
	// into the output, so we do the following:
	// We have to copy the result into a new buffer because the Fprintln over-writes
	// the result (what a pain).
	oldBuffer1 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer1 = append(oldBuffer1, bytesWritten[i])
	}
	o1 := bytes.NewBuffer(oldBuffer1)
	l := int64(o1.Len()) // cast to same type as returned by WriteTo()
	fmt.Fprintln(oldDestination, "Captured Event OLD 1:")
	if n, err := o1.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	oldBuffer2 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer2 = append(oldBuffer2, bytesWritten[i])
	}
	o2 := bytes.NewBuffer(oldBuffer2)
	l = int64(o2.Len())
	fmt.Fprintln(oldDestination, "Captured Event OLD 2:")
	if n, err := o2.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	// 3rd Event is like the second one in Middleware()
	// Capture the output of the call to Event()
	Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))

	oldBuffer3 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer3 = append(oldBuffer3, bytesWritten[i])
	}
	o3 := bytes.NewBuffer(oldBuffer3)
	l = int64(o3.Len())
	fmt.Fprintln(oldDestination, "Captured Event OLD 3:")
	if n, err := o3.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	//////////////////////
	// Capture NEW events

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	// 1st Event is like the first one in Middleware()
	Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

	newBuffer1 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer1 = append(newBuffer1, bytesWritten[i])
	}
	n1 := bytes.NewBuffer(newBuffer1)
	l = int64(n1.Len()) // cast to same type as returned by WriteTo()
	fmt.Fprintln(oldDestination, "Captured Event NEW 1:")
	if n, err := n1.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	newBuffer2 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer2 = append(newBuffer2, bytesWritten[i])
	}
	n2 := bytes.NewBuffer(newBuffer2)
	l = int64(n2.Len()) // cast to same type as returned by WriteTo()
	fmt.Fprintln(oldDestination, "Captured Event NEW 2:")
	if n, err := n2.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	// 3rd Event is like the second one in Middleware()
	// Capture the output of the call to Event()
	Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))

	newBuffer3 := make([]byte, 1)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer3 = append(newBuffer3, bytesWritten[i])
	}
	n3 := bytes.NewBuffer(newBuffer3)
	l = int64(n3.Len()) // cast to same type as returned by WriteTo()
	fmt.Fprintln(oldDestination, "Captured Event NEW 3:")
	if n, err := n3.WriteTo(oldDestination); n != l || err != nil {
		fmt.Println(err)
		return
	}

	//!!! add code to compare old and new events
}
