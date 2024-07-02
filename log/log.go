package log

import (
	"context"
	"github.com/ONSdigital/dp-net/v2/request"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12)
)

var logger *slog.Logger

// SetDefault makes l the default [slog.Logger] used by log.go, which is
// used by the top-level functions [Info], [Debug] and so on.
func SetDefault(l *slog.Logger) {
	logger = l
}

// Default returns the default [slog.Logger] used by log.go.
func Default() *slog.Logger {
	if logger == nil {
		return slog.Default()
	}
	return logger
}

// Event logs an event, to STDOUT if possible, or STDERR if not.
//
// Context can be nil.
//
// An event string should be static strings which do not use
// concatenation or Sprintf, e.g.
//
//	"connecting to database"
//
// rather than
//
//	"connecting to database: " + databaseURL
//
// Additional data should be stored using Data{}
//
// You can also pass in additional options which log extra event
// data, for example using the HTTP, Auth, Severity, Data and Error
// functions.
//
//	log.Event(nil, "connecting to database", log.Data{"url": databaseURL})
//
// If HUMAN_LOG environment variable is set to a true value (true, TRUE, 1)
// the log output will be syntax highlighted pretty printed JSON. Otherwise,
// the output is JSONLines format, with one JSON object per line.
func Event(ctx context.Context, event string, severity severity, opts ...option) {
	attrs := createEvent(ctx, event, severity, opts...).ToAttrs()
	Default().LogAttrs(ctx, SeverityToLevel(severity), event, attrs...)
}

// Info wraps the Event function with the severity level set to INFO
func Info(ctx context.Context, event string, opts ...option) {
	attrs := createEvent(ctx, event, INFO, opts...).ToAttrs()
	Default().LogAttrs(ctx, slog.LevelInfo, event, attrs...)
}

// Warn wraps the Event function with the severity level set to WARN
func Warn(ctx context.Context, event string, opts ...option) {
	attrs := createEvent(ctx, event, WARN, opts...).ToAttrs()
	Default().LogAttrs(ctx, slog.LevelWarn, event, attrs...)
}

// Error wraps the Event function with the severity level set to ERROR
func Error(ctx context.Context, event string, err error, opts ...option) {
	attrs := createEvent(ctx, event, ERROR, opts...).ToAttrs()
	attrs = append(attrs, slog.Any("error", err))
	Default().LogAttrs(ctx, slog.LevelError, event, attrs...)
}

// Fatal wraps the Event function with the severity level set to FATAL
func Fatal(ctx context.Context, event string, err error, opts ...option) {
	attrs := createEvent(ctx, event, FATAL, opts...).ToAttrs()
	attrs = append(attrs, slog.Any("error", err))
	Default().LogAttrs(ctx, LevelFatal, event, attrs...)
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
	// Optional fields
	TraceID  string    `json:"trace_id,omitempty"`
	SpanID   string    `json:"span_id,omitempty"`
	Severity *severity `json:"severity,omitempty"`

	// Optional nested data
	HTTP *EventHTTP `json:"http,omitempty"`
	Data *Data      `json:"data,omitempty"`

	// Error data
	Errors *EventErrors `json:"errors,omitempty"`
}

// ToAttrs creates a slice of slog.ToAttrs from the existing EventData
func (ed *EventData) ToAttrs() []slog.Attr {
	attrs := make([]slog.Attr, 0, 7)

	if ed.TraceID != "" {
		attrs = append(attrs, slog.String("trace_id", ed.TraceID))
	}

	if ed.SpanID != "" {
		attrs = append(attrs, slog.String("span_id", ed.SpanID))
	}

	if ed.Severity != nil {
		attrs = append(attrs, slog.Int("severity", int(*ed.Severity)))
	}

	if ed.HTTP != nil {
		attrs = append(attrs, slog.Any("http", *ed.HTTP))
	}

	if ed.Data != nil {
		attrs = append(attrs, slog.Any("data", *ed.Data))
	}

	return attrs
}

// createEvent creates a new event struct and attaches the options to it
func createEvent(ctx context.Context, event string, severity severity, opts ...option) *EventData {
	e := EventData{
		Severity: &severity,
	}

	if ctx != nil {
		e.TraceID = getRequestId(ctx)
	}

	otelTraceId := trace.SpanFromContext(ctx).SpanContext().TraceID()
	if otelTraceId.IsValid() {
		e.TraceID = otelTraceId.String()
	}

	// loop around each log option and call its attach method, which takes care
	// of the association with the EventData struct
	for _, o := range opts {
		o.attach(&e)
	}

	return &e
}

func getRequestId(ctx context.Context) string {
	requestID := ctx.Value(request.RequestIdKey)
	if requestID == nil {
		requestID = ctx.Value("request-id")
	}

	correlationID, ok := requestID.(string)
	if !ok {
		return ""
	}

	return correlationID
}
