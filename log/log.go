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

var namespace = os.Args[0]

var destination = os.Stdout
var fallbackDestination = os.Stderr

var isTestMode = flag.Lookup("test.v") != nil

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

// Event ...
func Event(ctx context.Context, event string, opts ...Loggable) {
	e := EventData{
		CreatedAt: time.Now(),
		Namespace: namespace,
	}

	if isTestMode {
		/*
			Test for the same arg being passed in multiple times

			Only happens when using `go test` to avoid the runtime
			overhead of doing type checks on every log event
		*/
		var optMap = make(map[string]struct{})
		for _, o := range opts {
			t := reflect.TypeOf(o)
			p := fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
			if _, ok := optMap[p]; ok {
				panic("can't pass in the same parameter type multiple times: " + p)
			}
			optMap[p] = struct{}{}
		}
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
