package log

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"
)

// Namespace is the log namespace included with every log event.
//
// It defaults to the application binary name, but this should typically
// be set to a more sensible name on application startup
var Namespace = os.Args[0]

var destination = os.Stdout
var fallbackDestination = os.Stderr

func init() {
	if flag.Lookup("test.v") != nil {
		Event = EventWithOptionsCheck
	}
}

// EventFunc is a function which handles log events
type eventFunc = func(ctx context.Context, event string, opts ...Loggable)

// Event ...
var Event = EventWithoutOptionsCheck

// Loggable ...
type Loggable interface {
	Attach(*EventData)
}

// EventData ...
type EventData struct {
	// Required fields
	CreatedAt time.Time `json:"created_at"`
	Namespace string    `json:"namespace"`
	Event     string    `json:"event"`

	// Optional fields
	TraceID  string    `json:"trace_id,omitempty"`
	SpanID   string    `json:"span_id,omitempty"`
	Severity *Severity `json:"severity,omitempty"`

	// Optional nested data
	HTTP *EventHTTP `json:"http,omitempty"`
	Auth *EventAuth `json:"auth,omitempty"`
	Data *Data      `json:"data,omitempty"`
}

// EventWithOptionsCheck is the event function used when running tests, and
// will panic if the same log option is passed in multiple times
//
// It is only used during tests because of the runtime performance overhead
func EventWithOptionsCheck(ctx context.Context, event string, opts ...Loggable) {
	var optMap = make(map[string]struct{})
	for _, o := range opts {
		t := reflect.TypeOf(o)
		p := fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
		if _, ok := optMap[p]; ok {
			panic("can't pass in the same parameter type multiple times: " + p)
		}
		optMap[p] = struct{}{}
	}

	Event(ctx, event, opts...)
}

// EventWithoutOptionsCheck is the event function used when we're not running tests
//
// It doesn't do any log options checks to minimise the runtime performance overhead
func EventWithoutOptionsCheck(ctx context.Context, event string, opts ...Loggable) {
	e := EventData{
		CreatedAt: time.Now(),
		Namespace: Namespace,
		Event:     event,
	}

	for _, o := range opts {
		o.Attach(&e)
	}

	b, err := json.Marshal(e)
	if err != nil {
		// TODO
		return
	}

	// try and write to stdout
	if n, err := fmt.Fprintln(destination, string(b)); n != len(b) || err != nil {
		// if that fails, try and write to stderr
		// not much point catching this error since there's not a lot else we can do
		fmt.Fprintln(fallbackDestination, string(b))
	}
}
