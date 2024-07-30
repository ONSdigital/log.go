package log

import (
	"context"
	"log/slog"
)

// ModifyingHandler implements a [slog.Handler] that wraps attrs in a 'data' group and adds a severity attribute
// translated from the log level of the record.
//
// This can be used with a default logger so that logging from third party libraries is translated into a format that
// matches the dp standard logging structure.
type ModifyingHandler struct {
	baseHandler slog.Handler
}

// compile-time type check
var _ slog.Handler = ModifyingHandler{}

// Enabled calls the underlying base handler's Enabled method
func (mh ModifyingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return mh.baseHandler.Enabled(ctx, level)
}

// Handle adds a severity [slog.Attr] to the record translated from the log level and then loops over the record's
// attrs, wrapping them within a `data` group. It then calls the underlying Handle method on the base handler
func (mh ModifyingHandler) Handle(ctx context.Context, record slog.Record) error {
	modifiedRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	severity := LevelToSeverity(record.Level)
	severityAttr := slog.Int("severity", int(severity))
	modifiedRecord.AddAttrs(severityAttr)
	record.Attrs(func(attr slog.Attr) bool {
		modifiedRecord.AddAttrs(slog.Group("data", attr))
		return true
	})
	return mh.baseHandler.Handle(ctx, modifiedRecord)
}

// WithGroup returns a new [ModifyingHandler] using the underlying base handler's WithGroup method
func (mh ModifyingHandler) WithGroup(name string) slog.Handler {
	return &ModifyingHandler{baseHandler: mh.baseHandler.WithGroup(name)}
}

// WithAttrs returns a new [ModifyingHandler] using the underlying base handler's WithAttrs method
func (mh ModifyingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ModifyingHandler{baseHandler: mh.baseHandler.WithAttrs(attrs)}
}
