package log

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ONSdigital/dp-net/v2/request"
	. "github.com/smartystreets/goconvey/convey"
)

// withRequestID sets the correlation id on the context
func withRequestID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, "request-id", correlationID)
}

func TestLog(t *testing.T) {
	t.Parallel()

	Convey("Function createEvent creates a new event", t, func() {
		Convey("Function createEvent sets the TraceID field to the request ID in the context", func() {
			ctx := withRequestID(context.Background(), "trace ID")
			evt := createEvent(ctx, INFO)
			So(evt.TraceID, ShouldEqual, "trace ID")

			ctx = withRequestID(context.Background(), "another ID")
			evt = createEvent(ctx, INFO)
			So(evt.TraceID, ShouldEqual, "another ID")
		})
	})
}

func TestGetRequestID(t *testing.T) {
	t.Parallel()

	Convey("Given context contains a request id key of type string", t, func() {
		testCtx := context.WithValue(context.Background(), "request-id", "test123")

		Convey("When I try to retrieve the request id from the context", func() {
			requestID := getRequestID(testCtx)

			Convey("Then the request id value is returned", func() {
				So(requestID, ShouldEqual, "test123")
			})
		})
	})

	Convey("Given context contains a request id key of type request.ContextKey", t, func() {
		testCtx := context.WithValue(context.Background(), request.RequestIdKey, "test321")

		Convey("When I try to retrieve the request id from the context", func() {
			requestID := getRequestID(testCtx)

			Convey("Then the request id value is returned", func() {
				So(requestID, ShouldEqual, "test321")
			})
		})
	})

	Convey("Given context contains does not contain a request id value", t, func() {
		testCtx := context.Background()

		Convey("When I try to retrieve the request id from the context", func() {
			requestID := getRequestID(testCtx)

			Convey("Then the request id value is returned", func() {
				So(requestID, ShouldBeEmpty)
			})
		})
	})
}

func TestToAttrs(t *testing.T) {
	t.Parallel()

	const (
		testTraceID   = "trace-id"
		testSpanID    = "span-id"
		testDataKey   = "key"
		testDataValue = "value"
	)
	var (
		testSeverity = INFO
	)

	Convey("Given some EventData", t, func() {
		eventHTTP := EventHTTP{}
		eventData := Data{testDataKey: testDataValue}

		ed := EventData{
			TraceID:  testTraceID,
			SpanID:   testSpanID,
			Severity: &testSeverity,
			HTTP:     &eventHTTP,
			Data:     &eventData,
		}

		Convey("When it is converted to log attributes", func() {
			attrs := ed.ToAttrs()
			So(attrs, ShouldNotBeNil)

			Convey("Then the values are in the attributes as expected", func() {
				So(attrs, ShouldHaveLength, 5)
				So(attrs[0].Key, ShouldEqual, "trace_id")
				So(attrs[0].Value.String(), ShouldEqual, testTraceID)
				So(attrs[1].Key, ShouldEqual, "span_id")
				So(attrs[1].Value.String(), ShouldEqual, testSpanID)
				So(attrs[2].Key, ShouldEqual, "severity")
				So(attrs[2].Value.Int64(), ShouldEqual, testSeverity)
				So(attrs[3].Key, ShouldEqual, "http")
				httpValue := attrs[3].Value
				So(httpValue, ShouldNotBeNil)
				So(httpValue.Any(), ShouldResemble, eventHTTP)
				So(attrs[4].Key, ShouldEqual, "data")
				dataValue := attrs[4].Value
				So(dataValue, ShouldNotBeNil)
				So(dataValue.Any(), ShouldResemble, eventData)
			})
		})
	})
}

func TestInfo(t *testing.T) {
	// Can't be parallel

	// Return default logger after run
	currentDefault := Default()
	defer SetDefault(currentDefault)

	ctx := context.Background()
	const (
		testEvent = "some event"
	)

	Convey("Given a mocked default logger", t, func() {
		mockHndlr := mockHandler{}
		logger := slog.New(&mockHndlr)
		SetDefault(logger)

		Convey("When we log a simple message", func() {
			mockHndlr.Reset()
			Info(ctx, testEvent)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelInfo)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, INFO)
			})
		})

		Convey("When we log a message with data", func() {
			mockHndlr.Reset()
			data := Data{"key": "value"}
			Info(ctx, testEvent, data)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelInfo)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, INFO)
				dataAttr := getAttrFromRecord(&record, "data")
				So(dataAttr, ShouldNotBeNil)
				So(dataAttr.Value.Any(), ShouldResemble, data)
			})
		})
	})
}

func TestWarn(t *testing.T) {
	// Can't be parallel

	// Return default logger after run
	currentDefault := Default()
	defer SetDefault(currentDefault)

	ctx := context.Background()
	const (
		testEvent = "some event"
	)

	Convey("Given a mocked default logger", t, func() {
		mockHndlr := mockHandler{}
		logger := slog.New(&mockHndlr)
		SetDefault(logger)

		Convey("When we log a simple message", func() {
			mockHndlr.Reset()
			Warn(ctx, testEvent)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelWarn)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, WARN)
			})
		})

		Convey("When we log a message with data", func() {
			mockHndlr.Reset()
			data := Data{"key": "value"}
			Warn(ctx, testEvent, data)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelWarn)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, WARN)
				dataAttr := getAttrFromRecord(&record, "data")
				So(dataAttr, ShouldNotBeNil)
				So(dataAttr.Value.Any(), ShouldResemble, data)
			})
		})
	})
}

func TestError(t *testing.T) {
	// Can't be parallel

	// Return default logger after run
	currentDefault := Default()
	defer SetDefault(currentDefault)

	ctx := context.Background()
	const (
		testEvent = "some event"
	)
	var (
		testError = errors.New("some error")
	)

	Convey("Given a mocked default logger", t, func() {
		mockHndlr := mockHandler{}
		logger := slog.New(&mockHndlr)
		SetDefault(logger)

		Convey("When we log a simple message", func() {
			mockHndlr.Reset()
			Error(ctx, testEvent, testError)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelError)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, ERROR)
				errorAttr := getAttrFromRecord(&record, "error")
				So(errorAttr.Value.Any(), ShouldEqual, testError)
			})
		})

		Convey("When we log a message with data", func() {
			mockHndlr.Reset()
			data := Data{"key": "value"}
			Error(ctx, testEvent, testError, data)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelError)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, ERROR)
				errorAttr := getAttrFromRecord(&record, "error")
				So(errorAttr.Value.Any(), ShouldEqual, testError)
				dataAttr := getAttrFromRecord(&record, "data")
				So(dataAttr, ShouldNotBeNil)
				So(dataAttr.Value.Any(), ShouldResemble, data)
			})
		})
	})
}

func TestFatal(t *testing.T) {
	// Can't be parallel

	// Return default logger after run
	currentDefault := Default()
	defer SetDefault(currentDefault)

	ctx := context.Background()
	const (
		testEvent = "some event"
	)
	var (
		testError = errors.New("some error")
	)

	Convey("Given a mocked default logger", t, func() {
		mockHndlr := mockHandler{}
		logger := slog.New(&mockHndlr)
		SetDefault(logger)

		Convey("When we log a simple message", func() {
			mockHndlr.Reset()
			Fatal(ctx, testEvent, testError)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelFatal)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, FATAL)
				errorAttr := getAttrFromRecord(&record, "error")
				So(errorAttr.Value.Any(), ShouldEqual, testError)
			})
		})

		Convey("When we log a message with data", func() {
			mockHndlr.Reset()
			data := Data{"key": "value"}
			Fatal(ctx, testEvent, testError, data)

			Convey("Then …", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
				record := mockHndlr.handeRecords[0]
				So(record.Message, ShouldResemble, testEvent)
				So(record.Level, ShouldEqual, LevelFatal)
				sevAttr := getAttrFromRecord(&record, "severity")
				So(sevAttr.Value.Int64(), ShouldEqual, FATAL)
				errorAttr := getAttrFromRecord(&record, "error")
				So(errorAttr.Value.Any(), ShouldEqual, testError)
				dataAttr := getAttrFromRecord(&record, "data")
				So(dataAttr, ShouldNotBeNil)
				So(dataAttr.Value.Any(), ShouldResemble, data)
			})
		})
	})
}

type mockHandler struct {
	handeRecords []slog.Record
}

var _ slog.Handler = &mockHandler{}

func (m *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *mockHandler) Handle(ctx context.Context, record slog.Record) error {
	m.handeRecords = append(m.handeRecords, record)
	return nil
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	return m
}

func (m *mockHandler) Reset() {
	m.handeRecords = []slog.Record{}
}

func getAttrFromRecord(r *slog.Record, key string) *slog.Attr {
	var foundAttr *slog.Attr
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == key {
			foundAttr = &attr
			return false
		}
		return true
	})
	return foundAttr
}
