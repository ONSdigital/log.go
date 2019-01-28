package log

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"
)

// Middleware implements the logger middleware and captures HTTP request data
func Middleware(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rc := &responseCapture{w, 0}
		start := time.Now()
		Event(req.Context(), "http request received", HTTP(req, 0, &start, nil))

		defer func() {
			end := time.Now()
			Event(req.Context(), "http request completed", HTTP(req, rc.statusCode, &start, &end))
		}()

		f.ServeHTTP(rc, req)
	})
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseCapture) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseCapture) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("log: response does not implement http.Hijacker")
}
