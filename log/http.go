package log

import (
	"net/http"
	"strconv"
	"time"
)

// EventHTTP ...
type EventHTTP struct {
	StatusCode int    `json:"status_code,omitempty"`
	Method     string `json:"method,omitempty"`

	// URL data
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Path   string `json:"path,omitempty"`
	Query  string `json:"query,omitempty"`

	// Timing data
	StartedAt time.Time     `json:"started_at,omitempty"`
	EndedAt   time.Time     `json:"ended_at,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
}

// Attach ...
func (l *EventHTTP) Attach(le *EventData) {
	le.HTTP = l
}

// HTTP ...
func HTTP(req *http.Request, statusCode int, startedAt time.Time, endedAt time.Time) Loggable {
	port := 0
	if p := req.URL.Port(); len(p) > 0 {
		port, _ = strconv.Atoi(p)
	}

	return &EventHTTP{
		StatusCode: statusCode,
		Method:     req.Method,

		Scheme: req.URL.Scheme,
		Host:   req.URL.Hostname(),
		Port:   port,
		Path:   req.URL.Path,
		Query:  req.URL.RawQuery,

		StartedAt: startedAt,
		EndedAt:   endedAt,
		Duration:  endedAt.Sub(startedAt),
	}
}
