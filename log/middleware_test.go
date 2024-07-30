package log

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type responseWriterWithoutHijacker struct {
	http.ResponseWriter
}

type responseWriter struct {
	http.ResponseWriter
	hijackCalled bool
	flushCalled  bool
	writeCalled  bool
}

func (r *responseWriter) WriteHeader(status int) {}
func (r *responseWriter) Write(b []byte) (int, error) {
	r.writeCalled = true
	return len(b), nil
}
func (r *responseWriter) Flush() {
	r.flushCalled = true
}
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.hijackCalled = true
	return nil, nil, nil
}

func TestResponseCapture(t *testing.T) {
	Convey("Type responseCapture implements http.ResponseWriter", t, func() {
		r := &responseCapture{&responseWriter{}, nil, 0}
		So(r, ShouldImplement, (*http.ResponseWriter)(nil))
	})

	Convey("Type responseCapture implements http.Flusher", t, func() {
		rw := &responseWriter{}
		r := &responseCapture{rw, nil, 0}
		So(r, ShouldImplement, (*http.Flusher)(nil))
		So(rw.flushCalled, ShouldBeFalse)
		http.Flusher(r).Flush()
		So(rw.flushCalled, ShouldBeTrue)
	})

	Convey("Type responseCapture implements http.Hijacker", t, func() {
		rw := &responseWriter{}
		r := &responseCapture{rw, nil, 0}
		So(r, ShouldImplement, (*http.Hijacker)(nil))
		So(rw.hijackCalled, ShouldBeFalse)
		_, _, err := http.Hijacker(r).Hijack()
		So(rw.hijackCalled, ShouldBeTrue)
		So(err, ShouldBeNil)

		Convey("Hijack returns an error if the inner http.ResponseWriter isn't a http.Hijacker", func() {
			rw := &responseWriterWithoutHijacker{}
			r := &responseCapture{rw, nil, 0}
			_, _, err := http.Hijacker(r).Hijack()
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Type responseCapture records the status code", t, func() {
		Convey("Type responseCapture records the status code when calling WriteHeader", func() {
			r := &responseCapture{&responseWriter{}, nil, 0}
			So(r.statusCode, ShouldBeNil)
			r.WriteHeader(501)
			So(r.statusCode, ShouldNotBeNil)
			So(*r.statusCode, ShouldEqual, 501)
		})

		Convey("Type responseCapture records the status code when skipping WriteHeader", func() {
			r := &responseCapture{&responseWriter{}, nil, 0}
			So(r.statusCode, ShouldBeNil)
			r.Write([]byte{})
			So(r.statusCode, ShouldNotBeNil)
			So(*r.statusCode, ShouldEqual, 200)
		})
	})

	Convey("Type responseCapture records the number of bytes written", t, func() {
		r := &responseCapture{&responseWriter{}, nil, 0}
		So(r.bytesWritten, ShouldEqual, 0)

		r.Write([]byte("abc"))
		So(r.bytesWritten, ShouldEqual, 3)

		r.Write([]byte("def"))
		So(r.bytesWritten, ShouldEqual, 6)
	})
}

func TestMiddleware(t *testing.T) {
	// Return default logger after run
	currentDefault := Default()
	defer SetDefault(currentDefault)

	Convey("Middleware wraps a http.Handler", t, func() {
		mockHndlr := mockHandler{}
		logger := slog.New(&mockHndlr)
		SetDefault(logger)

		var handlerWasCalled bool
		h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerWasCalled = true
			w.WriteHeader(200)
		})
		m := Middleware(h)
		So(m, ShouldHaveSameTypeAs, h)

		Convey("Middleware logs an event on nil request", func() {
			mockHndlr.Reset()

			m.ServeHTTP(nil, nil)

			So(mockHndlr.handeRecords, ShouldHaveLength, 1)
			record := mockHndlr.handeRecords[0]
			So(record.Message, ShouldResemble, "nil request in middleware handler")
			So(record.Level, ShouldEqual, LevelInfo)

			values := getValuesFromRecord(&record)
			So(values, ShouldNotBeEmpty)

			So(values, ShouldContainKey, "severity")
			value := values["severity"]
			So(value.Kind(), ShouldEqual, slog.KindInt64)
			So(value.Int64(), ShouldEqual, INFO)

			value = values["data"]
			So(value.Any(), ShouldNotBeNil)
			So(value.Any(), ShouldHaveSameTypeAs, Data{})
		})

		Convey("Inner handler is called by middleware", func() {
			So(handlerWasCalled, ShouldBeFalse)
			req, err := http.NewRequest("GET", "/", http.NoBody)
			So(err, ShouldBeNil)
			So(req, ShouldNotBeNil)
			m.ServeHTTP(&responseWriter{}, req)
			So(handlerWasCalled, ShouldBeTrue)
		})

		Convey("Start and end events are logged", func() {
			mockHndlr.Reset()

			req, err := http.NewRequest("GET", "http://localhost:1234/a/b/c?x=1&y=2", http.NoBody)
			So(err, ShouldBeNil)
			ctx := context.Background()
			req = req.WithContext(ctx)
			So(req, ShouldNotBeNil)
			m.ServeHTTP(&responseWriter{}, req)

			So(mockHndlr.handeRecords, ShouldHaveLength, 2)

			Convey("Start event is logged", func() {
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, "http request received")
				So(record.Level, ShouldEqual, LevelInfo)

				values := getValuesFromRecord(&record)
				So(values, ShouldNotBeEmpty)

				So(values, ShouldContainKey, "severity")
				value := values["severity"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, INFO)

				value = values["http"]
				So(value.Any(), ShouldNotBeNil)
				So(value.Any(), ShouldHaveSameTypeAs, EventHTTP{})

				eventHTTP := value.Any().(EventHTTP)
				So(eventHTTP.StatusCode, ShouldNotBeNil)
				So(*eventHTTP.StatusCode, ShouldEqual, 0)
				So(eventHTTP.Method, ShouldEqual, "GET")

				So(eventHTTP.Scheme, ShouldEqual, "http")
				So(eventHTTP.Host, ShouldEqual, "localhost")
				So(eventHTTP.Port, ShouldEqual, 1234)
				So(eventHTTP.Path, ShouldEqual, "/a/b/c")
				So(eventHTTP.Query, ShouldEqual, "x=1&y=2")

				// TODO more than nil check test for start/end times
				So(eventHTTP.StartedAt, ShouldNotBeNil)
				So(eventHTTP.EndedAt, ShouldBeNil)
				So(eventHTTP.Duration, ShouldBeNil)
				So(eventHTTP.ResponseContentLength, ShouldEqual, 0)
			})

			Convey("End event is logged", func() {
				record := mockHndlr.handeRecords[1]
				So(record.Message, ShouldResemble, "http request completed")
				So(record.Level, ShouldEqual, LevelInfo)

				values := getValuesFromRecord(&record)
				So(values, ShouldNotBeEmpty)

				So(values, ShouldContainKey, "severity")
				value := values["severity"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, INFO)

				value = values["http"]
				So(value.Any(), ShouldNotBeNil)
				So(value.Any(), ShouldHaveSameTypeAs, EventHTTP{})

				eventHTTP := value.Any().(EventHTTP)

				So(eventHTTP.StatusCode, ShouldNotBeNil)
				So(*eventHTTP.StatusCode, ShouldEqual, 200)
				So(eventHTTP.Method, ShouldEqual, "GET")

				So(eventHTTP.Scheme, ShouldEqual, "http")
				So(eventHTTP.Host, ShouldEqual, "localhost")
				So(eventHTTP.Port, ShouldEqual, 1234)
				So(eventHTTP.Path, ShouldEqual, "/a/b/c")
				So(eventHTTP.Query, ShouldEqual, "x=1&y=2")

				// TODO more than nil check test for start/end times
				So(eventHTTP.StartedAt, ShouldNotBeNil)
				So(eventHTTP.EndedAt, ShouldNotBeNil)
				So(eventHTTP.Duration, ShouldNotBeNil)
				So(eventHTTP.ResponseContentLength, ShouldEqual, 0)
			})
		})
	})
}
