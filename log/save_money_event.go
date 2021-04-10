package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/common"
)

// Each SaveMoneyEvent... need their own sync.Pool
var eventBufPool1 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

var eventBufPool2 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

var eventBufPool3 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

// SaveMoneyEvent1 for use in middleware to replace the 1st log.Event
func SaveMoneyEvent1(ctx context.Context, event string, opts ...option) {
	// Minimum Allocations Event code ...
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

	var somethingWritten bool

	buf := eventBufPool1.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	if !e.CreatedAt.IsZero() {
		somethingWritten = true
		unrollCreatedAt(buf, e.CreatedAt)
	}

	if e.Namespace != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollNamespace(buf, e.Namespace)
	}

	if e.Event != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollEvent(buf, e.Event)
	}

	if e.TraceID != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollTraceID(buf, e.TraceID)
	}

	if e.Severity != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollSeverity(buf, int(*e.Severity))
	}

	if e.HTTP != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollHTTPToBuf(buf, e.HTTP, true, false, false, false, true, true)
	}

	if e.Auth != nil {
		unrollAuthToBuf(somethingWritten, buf, e.Auth)
		somethingWritten = true
	}

	if e.Data != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollDataToBuf(buf, e.Data)
	}

	if e.Error != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		unrollErrorToBuf(buf, e.Error)
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

	eventBufPool1.Put(buf)
}

// SaveMoneyEvent2 for use in dp-frontend-router to replace log.Event in
// function: createReverseProxy()
func SaveMoneyEvent2(ctx context.Context, event string, opts ...option) {
	// Minimum Allocations Event code ...
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

	var somethingWritten bool

	buf := eventBufPool2.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	if !e.CreatedAt.IsZero() {
		somethingWritten = true
		unrollCreatedAt(buf, e.CreatedAt)
	}

	if e.Namespace != "" {
		/*if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollNamespace(buf, e.Namespace)*/
	}

	if e.Event != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollEvent(buf, e.Event)
	}

	if e.TraceID != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollTraceID(buf, e.TraceID)
	}

	if e.Severity != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollSeverity(buf, int(*e.Severity))
	}

	if e.HTTP != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		// keep query
		unrollHTTPToBuf(buf, e.HTTP, false, false, false, false, false, true)
	}

	if e.Auth != nil {
		unrollAuthToBuf(somethingWritten, buf, e.Auth)
		somethingWritten = true
	}

	if e.Data != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollDataToBuf(buf, e.Data)
	}

	if e.Error != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		unrollErrorToBuf(buf, e.Error)
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

	eventBufPool2.Put(buf)
}

// SaveMoneyEvent3 for use in middleware to replace the 2nd log.Event
func SaveMoneyEvent3(ctx context.Context, event string, opts ...option) {
	// Minimum Allocations Event code ...
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

	var somethingWritten bool

	buf := eventBufPool3.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	if !e.CreatedAt.IsZero() {
		somethingWritten = true
		unrollCreatedAt(buf, e.CreatedAt)
	}

	if e.Namespace != "" {
		/*if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollNamespace(buf, e.Namespace)*/
	}

	if e.Event != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollEvent(buf, e.Event)
	}

	if e.TraceID != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollTraceID(buf, e.TraceID)
	}

	if e.Severity != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollSeverity(buf, int(*e.Severity))
	}

	if e.HTTP != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollHTTPToBuf(buf, e.HTTP, false, false, false, false, false, false)
	}

	if e.Auth != nil {
		unrollAuthToBuf(somethingWritten, buf, e.Auth)
		somethingWritten = true
	}

	if e.Data != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		unrollDataToBuf(buf, e.Data)
	}

	if e.Error != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		unrollErrorToBuf(buf, e.Error)
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

	eventBufPool3.Put(buf)
}
