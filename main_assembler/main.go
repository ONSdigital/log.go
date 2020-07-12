package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"reflect"
	"runtime"

	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/ONSdigital/go-ns/common"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

// compile this with:
// go tool compile -S -I $GOPATH/pkg/linux_amd64 main.go >main.asm
//
// to then inspect assembler code in "main.asm"

/* 2nd May 2020 :

Observations of assembler code differences for o.attach(&e) compared to switch()
in functions: 	BenchmarkLog2() and BenchmarkLog3()  respectively.
---------------------------------------------------------------------------------

The original .go code has FIVE attach() methods:

func (l *eventAuth) attach(le *EventData) {
func (d Data) attach(le *EventData) {
func (l *EventError) attach(le *EventData) {
func (l *EventHTTP) attach(le *EventData) {
func (s severity) attach(le *EventData) {

The assembler code for these original functions has EIGHT chunks of code functions:

"".option.attach STEXT dupok size=100 args=0x18 locals=0x18

"".(*Data).attach STEXT dupok size=183 args=0x10 locals=0x18
"".Data.attach STEXT size=136 args=0x10 locals=0x18

"".(*eventAuth).attach STEXT size=84 args=0x10 locals=0x8

"".(*EventError).attach STEXT size=84 args=0x10 locals=0x8

"".(*EventHTTP).attach STEXT size=84 args=0x10 locals=0x8

"".(*severity).attach STEXT dupok size=153 args=0x10 locals=0x18
"".severity.attach STEXT size=106 args=0x10 locals=0x18

-=-=-
TWO different ones for Data & severity ... might suggest that the switch() code i wrote needs two more case's ... hmmm

And i got no idea what the first "".option.attach is, or how one might us it ... it must come about from the definition of option in log.go :

type option interface {
	attach(*EventData)
}


*/

func main() {
	BenchmarkLog2()
	BenchmarkLog3()
}

func BenchmarkLog2() {
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

	// loop around each log option and call its attach method, which takes care
	// of the association with the EventData struct
	for _, o := range opts {
		// Using rare pattern : `thing.attach(toObject)`
		// this handles both cases where:
		// the receiver can be called either `dataThing.attach(...)` or `ptrToDataThing.attach(...)
		o.attach(&e)
	}
}

func BenchmarkLog3() {
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

//////////// pull in code from log.go to get compile with assembler output to play ball

// from : auth.go  ////////////////////////////

type eventAuth struct {
	Identity     string       `json:"identity,omitempty"`
	IdentityType identityType `json:"identity_type,omitempty"`
}

type identityType string

const (
	// SERVICE represents a service account type
	SERVICE identityType = "service"
	// USER represents a user account type
	USER identityType = "user"
)

func (l *eventAuth) attach(le *EventData) {
	le.Auth = l
}

// Auth returns an option you can pass to Event to include identity information,
// for example the identity type and user/service ID from an inbound HTTP request
func Auth(identityType identityType, identity string) option {
	return &eventAuth{
		Identity:     identity,
		IdentityType: identityType,
	}
}

// from : data.go  ////////////////////////////

// Data can be used to include arbitrary key/value pairs
// in the structured log output.
//
// This should only be used where a predefined field isn't
// already available, since data included in a Data{} value
// isn't easily indexable.
//
// You can also create nested log data, for example:
//     Data {
//          "key": Data{},
//     }
type Data map[string]interface{}

func (d Data) attach(le *EventData) {
	le.Data = &d
}

// from : error.go  ////////////////////////////

// EventError is the data structure used for logging a error event.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventError struct {
	Error      string            `json:"error,omitempty"`
	StackTrace []EventStackTrace `json:"stack_trace,omitempty"`
	// This uses interface{} type, but should always be a type of kind struct
	// (which serialises to map[string]interface{})
	// See `func Error` switch block for more info
	Data interface{} `json:"data,omitempty"`
}

// EventStackTrace is the data structure used for logging a stack trace.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventStackTrace struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

func (l *EventError) attach(le *EventData) {
	le.Error = l
}

// Error returns an option you can pass to Event to attach
// error information to a log event
//
// It uses error.Error() to stringify the error value
//
// It also includes the error type itself as unstructured log
// data. For a struct{} type, it is included directly. For all
// other types, it is wrapped in a Data{} struct
//
// It also includes a full strack trace to where Error() is called,
// so you shouldn't normally store a log.Error for reuse (e.g. as a
// package level variable)
func Error(err error) option {
	e := &EventError{
		Error:      err.Error(),
		StackTrace: make([]EventStackTrace, 0),
	}

	k := reflect.Indirect(reflect.ValueOf(err)).Type().Kind()
	switch k {
	case reflect.Struct:
		// We've got a struct type, so make it the top level value
		e.Data = err
	default:
		// We have something else, so nest it inside a Data value
		e.Data = Data{"value": err}
	}

	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	if n > 0 {
		frames := runtime.CallersFrames(pc[:n])

		for {
			frame, more := frames.Next()

			e.StackTrace = append(e.StackTrace, EventStackTrace{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
			})

			if !more {
				break
			}
		}
	}

	return e
}

// from : http.go  //////////////////////////////

// EventHTTP is the data structure used for logging a HTTP event.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventHTTP struct {
	StatusCode *int   `json:"status_code,omitempty"`
	Method     string `json:"method,omitempty"`

	// URL data
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Path   string `json:"path,omitempty"`
	Query  string `json:"query,omitempty"`

	// Timing data
	StartedAt             *time.Time     `json:"started_at,omitempty"`
	EndedAt               *time.Time     `json:"ended_at,omitempty"`
	Duration              *time.Duration `json:"duration,omitempty"`
	ResponseContentLength int64          `json:"response_content_length,omitempty"`
}

func (l *EventHTTP) attach(le *EventData) {
	le.HTTP = l
}

// HTTP returns an option you can pass to Event to log HTTP
// request data with a log event.
//
// It converts the port number to a integer if possible, otherwise
// the port number is 0.
//
// It splits the URL into its component parts, and stores the scheme,
// host, port, path and query string individually.
//
// It also calculates the duration if both startedAt and endedAt are
// passed in, for example when wrapping a http.Handler.
func HTTP(req *http.Request, statusCode int, responseContentLength int64, startedAt, endedAt *time.Time) option {
	port := 0
	if p := req.URL.Port(); len(p) > 0 {
		port, _ = strconv.Atoi(p)
	}

	var duration *time.Duration
	if startedAt != nil && endedAt != nil {
		d := endedAt.Sub(*startedAt)
		duration = &d
	}

	return &EventHTTP{
		StatusCode: &statusCode,
		Method:     req.Method,

		Scheme: req.URL.Scheme,
		Host:   req.URL.Hostname(),
		Port:   port,
		Path:   req.URL.Path,
		Query:  req.URL.RawQuery,

		StartedAt:             startedAt,
		EndedAt:               endedAt,
		Duration:              duration,
		ResponseContentLength: responseContentLength,
	}
}

// from : log.go  /////////////////////////

// Namespace is the log namespace included with every log event.
//
// It defaults to the application binary name, but this should
// normally be set to a more sensible name on application startup
var Namespace = os.Args[0]

var destination io.Writer = os.Stdout
var fallbackDestination io.Writer = os.Stderr

var isTestMode bool

var eventWithOptionsCheckFunc = &eventFunc{eventWithOptionsCheck}
var eventWithoutOptionsCheckFunc = &eventFunc{eventWithoutOptionsCheck}
var eventFuncInst = initEvent()

var styleForHumanFunc = &styleFunc{styleForHuman}
var styleForMachineFunc = &styleFunc{styleForMachine}

// Event logs an event, to STDOUT if possible, or STDERR if not.
//
// Context can be nil.
//
// An event string should be static strings which do not use
// concatenation or Sprintf, e.g.
//     "connecting to database"
// rather than
//     "connecting to database: " + databaseURL
//
// Additional data should be stored using Data{}
//
// You can also pass in additional options which log extra event
// data, for example using the HTTP, Auth, Severity, Data and Error
// functions.
//
//     log.Event(nil, "connecting to database", log.Data{"url": databaseURL})
//
// If HUMAN_LOG environment variable is set to a true value (true, TRUE, 1)
// the log output will be syntax highlighted pretty printed JSON. Otherwise,
// the output is JSONLines format, with one JSON object per line.
//
// When running tests, Event will panic if the same option is passed
// in multiple times, for example:
//
//     log.Event(nil, "event", log.Data{}, log.Data{})
//
// It doesn't panic in normal usage because checking for duplicate entries
// is expensive. Where this happens, options to the right take precedence,
// for example:
//
//     log.Event(nil, "event", log.Data{"a": 1}, log.Data{"a": 2})
//     // data.a = 2
//
func Event(ctx context.Context, event string, opts ...option) {
	eventFuncInst.f(ctx, event, opts...)
}

func initEvent() *eventFunc {
	// If we're in test mode, replace the Event function with one
	// that has additional checks to find repeated event option types
	//
	// In test mode, a log event like this will result in a panic:
	//
	//    log.Event(nil, "demo", log.FATAL, log.WARN, log.ERROR)
	//
	// A flag called `test.v` is added by `go test`, so we can rely
	// on that to detect test mode.
	if flag.Lookup("test.v") != nil {
		isTestMode = true
		return eventWithOptionsCheckFunc
	}

	isTestMode = false
	return eventWithoutOptionsCheckFunc
}

var styler = initStyler()

func initStyler() *styleFunc {
	// If HUMAN_LOG is enabled, replace the default styler with a
	// human readable styler
	if b, _ := strconv.ParseBool(os.Getenv("HUMAN_LOG")); b {
		return styleForHumanFunc
	}

	return styleForMachineFunc
}

// eventFunc is a function which handles log events
type eventFunc struct {
	f func(ctx context.Context, event string, opts ...option)
}
type styleFunc = struct {
	f func(ctx context.Context, e EventData, ef eventFunc) []byte
}

// option is the interface which log options passed to eventFunc must match
//
// there's no point exporting this since it would require changes to the
// EventData struct (unless it forces data into log.Data or some other field,
// but we probably don't want that)
type option interface {
	attach(*EventData)
}

// EventData is the data structure used for logging an event
//
// It is the top level structure which contains all other log event data.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventData struct {
	// Required fields
	CreatedAt time.Time `json:"created_at"`
	Namespace string    `json:"namespace"`
	Event     string    `json:"event"`

	// Optional fields
	TraceID  string    `json:"trace_id,omitempty"`
	SpanID   string    `json:"span_id,omitempty"`
	Severity *severity `json:"severity,omitempty"`

	// Optional nested data
	HTTP *EventHTTP `json:"http,omitempty"`
	Auth *eventAuth `json:"auth,omitempty"`
	Data *Data      `json:"data,omitempty"`

	// Error data
	Error *EventError `json:"error,omitempty"`
}

// eventWithOptionsCheck is the event function used when running tests, and
// will panic if the same log option is passed in multiple times
//
// It is only used during tests because of the runtime performance overhead
func eventWithOptionsCheck(ctx context.Context, event string, opts ...option) {
	var optMap = make(map[string]struct{})
	for _, o := range opts {
		t := reflect.TypeOf(o)
		p := fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
		if _, ok := optMap[p]; ok {
			panic("can't pass in the same parameter type multiple times: " + p)
		}
		optMap[p] = struct{}{}
	}

	eventWithoutOptionsCheckFunc.f(ctx, event, opts...)
}

var output []byte

// eventWithoutOptionsCheck is the event function used when we're not running tests
//
// It doesn't do any log options checks to minimise the runtime performance overhead
func eventWithoutOptionsCheck(ctx context.Context, event string, opts ...option) {
	//	output = styler.f(ctx, *createEvent(ctx, event, opts...), eventFunc{eventWithoutOptionsCheck})
	print(styler.f(ctx, *createEvent(ctx, event, opts...), eventFunc{eventWithoutOptionsCheck}))
	//	print(output)
}

// createEvent creates a new event struct and attaches the options to it
func createEvent(ctx context.Context, event string, opts ...option) *EventData {
	e := EventData{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     event,
	}

	if ctx != nil {
		e.TraceID = common.GetRequestId(ctx)
	}

	// loop around each log option and call its attach method, which takes care
	// of the association with the EventData struct
	for _, o := range opts {
		//		o.attach(&e)
		switch v := o.(type) { // OR do assignments directly ...
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
			//		default:
			//			fmt.Printf("option: %v, %v, %T", o, v, v)
			//			panic("unknown option")
		}
	}

	return &e
}

// handleStyleError handles any errors from JSON marshalling in one of the styler functions
func handleStyleError(ctx context.Context, e EventData, ef eventFunc, b []byte, err error) []byte {
	if err != nil {
		// marshalling failed, so we'll log a marshalling error and use Sprintf
		// to get some kind of text representation of the log data
		//
		// other than out of memory errors, marshalling can only fail for an unsupported type
		// e.g. using log.Data and passing in an io.Reader
		//
		// to avoid this becoming recursive, only pass primitive types in this line (string, int, etc)
		//
		// note: Error(err) currently ignores this constraint, but it's expected that the `err`
		// 		 passed in by the caller will have come from json.Marshal or prettyjson.Marshal
		//       which don't marshal any non-marshallable types anyway
		ef.f(ctx, "error marshalling event data", Error(err), Data{"event_data": fmt.Sprintf("%+v", e)})

		// if we're in test mode, we'll also panic to cause tests to fail
		if isTestMode {
			// don't capture and reuse fmt.Sprintf output above for this, since that adds
			// a performance/memory overhead, and reuse is only required in test mode
			panic("error marshalling event data: " + fmt.Sprintf("%+v", e))
		}

		return []byte{}
	}

	return b
}

// styleForMachine renders the event data in JSONLine format
func styleForMachine(ctx context.Context, e EventData, ef eventFunc) []byte {
	b, err := json.Marshal(e)

	return handleStyleError(ctx, e, ef, b, err)
}

// styleForHuman renders the event data in a human readable format
func styleForHuman(ctx context.Context, e EventData, ef eventFunc) []byte {
	b, err := prettyjson.Marshal(e)

	return handleStyleError(ctx, e, ef, b, err)
}

func print(b []byte) {
	if len(b) == 0 {
		return
	}

	// try and write to stdout
	if n, err := fmt.Fprintln(destination, string(b)); n != len(b)+1 || err != nil {
		// if that fails, try and write to stderr
		if n, err := fmt.Fprintln(fallbackDestination, string(b)); n != len(b)+1 || err != nil {
			// if that fails, panic!
			//
			// also defer an os.Exit since the panic might be captured in a recover
			// block in the caller, but we always want to exit in this scenario
			//
			// Note: deferring an os.Exit makes this particular block untestable
			// using conventional `go test`. But it's a narrow enough edge case that
			// it probably isn't worth trying, and only occurs in extreme circumstances
			// (os.Stdout and os.Stderr both being closed) where unpredictable
			// behaviour is expected. It's not clear what a panic or os.Exit would do
			// in this scenario, or if our process is even still alive to get this far.
			defer os.Exit(1)
			panic("error writing log data: " + err.Error())
		}
	}
}

// from : severity.go  /////////////////////

const (
	// FATAL is an option you can pass to Event to specify a severity of FATAL/0
	FATAL severity = 0
	// ERROR is an option you can pass to Event to specify a severity of ERROR/1
	ERROR severity = 1
	// WARN is an option you can pass to Event to specify a severity of WARN/2
	WARN severity = 2
	// INFO is an option you can pass to Event to specify a severity of INFO/3
	INFO severity = 3
)

// severity is the log severity level
//
// we don't export this because we don't want the caller
// to define their own severity levels
type severity int

func (s severity) attach(le *EventData) {
	le.Severity = &s
}
