package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var namespace = os.Args[0]
var destination = os.Stdout

// Loggable ...
type Loggable interface {
	attach(*logEvent)
}

type logEvent struct {
	CreatedAt time.Time     `json:"created_at"`
	Namespace string        `json:"namespace"`
	HTTP      *logEventHTTP `json:"http,omitempty"`
	Data      *Data         `json:"data,omitempty"`
}

// Data ...
type Data map[string]interface{}

func (d Data) attach(le *logEvent) {
	le.Data = &d
}

type logEventHTTP struct {
	StatusCode int           `json:"status_code"`
	StartedAt  time.Time     `json:"started_at"`
	EndedAt    time.Time     `json:"ended_at"`
	Duration   time.Duration `json:"duration"`
	Path       string        `json:"path"`
	Host       string        `json:"host"`
	Port       int           `json:"port"`
	Query      string        `json:"query"`
}

func (l *logEventHTTP) attach(le *logEvent) {
	le.HTTP = l
}

type httpLogData struct {
	req *http.Request
}

// HTTP ...
func HTTP(req *http.Request, statusCode int, startedAt time.Time, endedAt time.Time) Loggable {
	return &logEventHTTP{}
}

// Event ...
func Event(event string, opts ...Loggable) {
	e := logEvent{
		CreatedAt: time.Now(),
		Namespace: namespace,
	}

	for _, o := range opts {
		o.attach(&e)
	}

	b, err := json.Marshal(e)
	if err != nil {
		// TODO
	}

	n, err := fmt.Fprintln(destination, string(b))
	if err != nil {
		// TODO
	}
	if n != len(b) {
		// TODO
	}
}
