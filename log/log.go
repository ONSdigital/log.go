package log

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/common"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

// Namespace is the log namespace included with every log event.
//
// It defaults to the application binary name, but this should
// normally be set to a more sensible name on application startup
var Namespace = os.Args[0]

var destination io.Writer = os.Stdout
var fallbackDestination io.Writer = os.Stderr

var isTestMode bool
var isMinimalAllocations bool

var eventWithOptionsCheckFunc = &eventFunc{eventWithOptionsCheck}
var eventWithoutOptionsCheckFunc = &eventFunc{eventWithoutOptionsCheck}
var eventFuncInst = initEvent()

var styleForHumanFunc = &styleFunc{styleForHuman}
var styleForMachineFunc = &styleFunc{styleForMachine}

var bufPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

func expandIntToBuf2(buf *bytes.Buffer, value int) {
	c1 := byte((value % 10) + '0')
	c2 := byte(((value / 10) % 10) + '0')
	buf.WriteByte(c2)
	buf.WriteByte(c1)
}

func expandIntToBuf4(buf *bytes.Buffer, value int) {
	c1 := byte((value % 10) + '0')
	value = value / 10
	c2 := byte((value % 10) + '0')
	value = value / 10
	c3 := byte((value % 10) + '0')
	value = value / 10
	c4 := byte((value % 10) + '0')
	buf.WriteByte(c4)
	buf.WriteByte(c3)
	buf.WriteByte(c2)
	buf.WriteByte(c1)
}

func expandIntToBuf9(buf *bytes.Buffer, value int) {
	var out [11]byte

	//value2 := value

	out[8] = byte((value % 10) + '0')
	value = value / 10
	out[7] = byte((value % 10) + '0')
	value = value / 10
	out[6] = byte((value % 10) + '0')
	value = value / 10
	out[5] = byte((value % 10) + '0')
	value = value / 10
	out[4] = byte((value % 10) + '0')
	value = value / 10
	out[3] = byte((value % 10) + '0')
	value = value / 10
	out[2] = byte((value % 10) + '0')
	value = value / 10
	out[1] = byte((value % 10) + '0')
	value = value / 10
	out[0] = byte((value % 10) + '0')

	var last int = 0
	for i := 8; i >= 0; i-- {
		if out[i] != '0' {
			last = i
			break
		}
	}
	// now output all digits, except for any trailing zero's
	for i := 0; i <= last; i++ {
		buf.WriteByte(out[i])
	}
	buf.WriteByte('Z')
}

func expandTimeToBuf(buf *bytes.Buffer, value time.Time) {
	expandIntToBuf4(buf, value.Year())
	buf.WriteByte('-')
	expandIntToBuf2(buf, int(value.Month()))
	buf.WriteByte('-')
	expandIntToBuf2(buf, int(value.Day()))
	buf.WriteByte('T')
	expandIntToBuf2(buf, int(value.Hour()))
	buf.WriteByte(':')
	expandIntToBuf2(buf, int(value.Minute()))
	buf.WriteByte(':')
	expandIntToBuf2(buf, int(value.Second()))
	buf.WriteByte('.')
	expandIntToBuf9(buf, int(value.Nanosecond()))
}

func expandInt(buf *bytes.Buffer, n int) {
	var out [15]byte
	var c int

	if n < 0 {
		n = -n
		buf.WriteByte('-')
	}
	for {
		out[c] = byte((n % 10) + '0')
		c++
		n = n / 10
		if n == 0 {
			break
		}
	}

	c--
	for {
		buf.WriteByte(out[c])
		if c == 0 {
			break
		}
		c--
	}
}

func expandInt64(buf *bytes.Buffer, n int64) {
	var out [25]byte
	var c int

	if n < 0 {
		n = -n
		buf.WriteByte('-')
	}
	for {
		out[c] = byte((n % 10) + '0')
		c++
		n = n / 10
		if n == 0 {
			break
		}
	}

	c--
	for {
		buf.WriteByte(out[c])
		if c == 0 {
			break
		}
		c--
	}
}

func expandHTTPToBuf(buf *bytes.Buffer, value *EventHTTP) {
	// We know what the '*EventHTTP' is, so its contents can be directly
	// extracted ...
	buf.WriteByte('{')

	var somethingWritten bool
	if value.StatusCode != nil {
		buf.WriteByte('"')
		buf.WriteString("status_code")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandInt(buf, *value.StatusCode)
		somethingWritten = true
	}

	if value.Method != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("method")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Method)
		buf.WriteByte('"')
	}

	if value.Scheme != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("scheme")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Scheme)
		buf.WriteByte('"')
	}

	if value.Host != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("host")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Host)
		buf.WriteByte('"')
	}

	// port will always have some value ...
	if somethingWritten {
		buf.WriteByte(',')
	} else {
		somethingWritten = true
	}
	buf.WriteByte('"')
	buf.WriteString("port")
	buf.WriteByte('"')
	buf.WriteByte(':')
	expandInt(buf, value.Port)

	if value.Path != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("path")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Path)
		buf.WriteByte('"')
	}

	if value.Query != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("query")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Query)
		buf.WriteByte('"')
	}

	if value.StartedAt != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("started_at")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		expandTimeToBuf(buf, *value.StartedAt)
		buf.WriteByte('"')
	}

	if value.EndedAt != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("ended_at")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		expandTimeToBuf(buf, *value.EndedAt)
		buf.WriteByte('"')
	}

	if value.Duration != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("duration")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandInt64(buf, int64(*value.Duration))
	}

	// We can not easily determine if ResponseContentLength has been assigned a value
	// without doing some sort of reflect which will then create allocs.
	// But if ResponseContentLength is 0, then there is no point showing it.
	if value.ResponseContentLength != 0 {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("response_content_length")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandInt64(buf, int64(value.ResponseContentLength))
	}

	buf.WriteByte('}')
}

//!!! expand Auth struct
func expandAuthToBuf(buf *bytes.Buffer, value *eventAuth) {
	json.NewEncoder(buf).Encode(value)
	buf.Truncate(buf.Len() - 1) // remove the 'new line', as there is more to append
}

func expandDataToBuf(buf *bytes.Buffer, value *Data) {
	// We know what the '*Data' is 'map[string]interface{}'
	// and we know that the three Event() calls on the HOT-PATH in dp-frontend-router
	// only pass a string or URL for the value ... so with this prior knowledge ...
	// to achieve the goal of minimum memory allocations, this function will
	// decode our 'knowns' in an allocation optimum way.
	// It will deal with unknowns with an inefficient "json.NewEncoder(buf).Encode(value)"
	// ... that said, it appears that the Encode does not create any allocations on
	// the HEAP for the URL (possibly because it has no sub structure structures or
	// interface{} ?)
	// ODD'ly: sometimes the 'proxy_name' is output before the 'destination' - go figure ?
	var somethingWritten bool
	buf.WriteByte('{')
	for k, v := range *value {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}

		buf.WriteByte('"')
		buf.WriteString(k) // add the key
		buf.WriteByte('"')
		buf.WriteByte(':')
		switch n := v.(type) {
		case string:
			buf.WriteByte('"')
			buf.WriteString(n) // add the 'known' value type of 'string'
			buf.WriteByte('"')
		default: // too many other possibilities, so use Encode()
			json.NewEncoder(buf).Encode(v)
			buf.Truncate(buf.Len() - 1) // remove the 'new line', as there is more to append
		}
	}
	buf.WriteByte('}')
}

//!!! expand Error struct
func expandErrorToBuf(buf *bytes.Buffer, value *EventError) {
	json.NewEncoder(buf).Encode(value)
	buf.Truncate(buf.Len() - 1) // remove the 'new line', as there is more to append
}

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
	if isMinimalAllocations == false {
		eventFuncInst.f(ctx, event, opts...)
		return
	}

	//fmt.Fprintln(destination, "SPEEEEED ...")
	e := EventData2{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     event,
	}

	if ctx != nil {
		e.TraceID = common.GetRequestId(ctx)
	}

	// loop around each log option and attach each option
	// directly into EventData2 struct
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

	//fmt.Fprintf(destination, "%+v\n", e)

	//var err error = nil

	//err := json.NewEncoder(destination).Encode(e)

	// The following is an 'inline' unrolling of:
	//    err := json.NewEncoder(destination).Encode(e)
	// to eliminate allocations leaking to the HEAP by using a
	// sync.Pool bytes.Buffer

	var somethingWritten bool

	buf := bufPool.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                          // Must reset before each block of usage

	buf.WriteByte('{')
	if !e.CreatedAt.IsZero() {
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("created_at")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		expandTimeToBuf(buf, e.CreatedAt)
		buf.WriteByte('"')
	}

	if e.Namespace != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("namespace")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(e.Namespace)
		buf.WriteByte('"')
	}

	if e.Event != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("event")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(e.Event)
		buf.WriteByte('"')
	}

	if e.TraceID != "" {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("trace_id")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(e.TraceID)
		buf.WriteByte('"')
	}

	if e.Severity != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("severity")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandInt(buf, int(*e.Severity))
	}

	if e.HTTP != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("http")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandHTTPToBuf(buf, e.HTTP)
	}

	//!!! run benchmark 6 to see this
	if e.Auth != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("auth")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandAuthToBuf(buf, e.Auth)
	}

	if e.Data != nil {
		if somethingWritten {
			buf.WriteByte(',')
		} else {
			somethingWritten = true
		}
		buf.WriteByte('"')
		buf.WriteString("data")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandDataToBuf(buf, e.Data)
	}

	//!!! run benchmark 6 to see this
	if e.Error != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString("error")
		buf.WriteByte('"')
		buf.WriteByte(':')
		expandErrorToBuf(buf, e.Error)
	}

	buf.WriteByte('}')
	buf.WriteByte(10)

	l := int64(buf.Len()) // cast to same type as returned by WriteTo()

	// try and write to stdout
	if n, err := buf.WriteTo(destination); n != l || err != nil {
		// if that fails, try and write to stderr
		if n, err := buf.WriteTo(fallbackDestination); n != l || err != nil {
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

	bufPool.Put(buf)
}

// this is called before main()
func initEvent() *eventFunc {
	if flag.Lookup("minimumAllocs") != nil {
		isMinimalAllocations = true
		//fmt.Printf("\n\nmin alloccs\n\n")
		//os.Exit(2)
	}
	if b, _ := strconv.ParseBool(os.Getenv("MINIMUM_ALLOC")); b {
		isMinimalAllocations = true
		//fmt.Printf("\n\nmin alloccs M\n\n")
		//os.Exit(3)
	}

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

// this is called before main()
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

// EventData2 - this version of 'EventData' has "SpanID" removed to reduce memory allocation in the HOT-PATH
type EventData2 struct {
	// Required fields
	CreatedAt time.Time `json:"created_at"`
	Namespace string    `json:"namespace"`
	Event     string    `json:"event"`

	// Optional fields
	TraceID  string    `json:"trace_id,omitempty"`
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

// eventWithoutOptionsCheck is the event function used when we're not running tests
//
// It doesn't do any log options checks to minimise the runtime performance overhead
func eventWithoutOptionsCheck(ctx context.Context, event string, opts ...option) {
	print(styler.f(ctx, *createEvent(ctx, event, opts...), eventFunc{eventWithoutOptionsCheck}))
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
		o.attach(&e)
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

	//	b = append(b, 55) // used to break test, just to check that test is working !!! (remove when memory alllocation optimisation finished)
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
