package log

import "log/slog"

const (
	// FATAL is an option you can pass to Event to specify a severity of FATAL/0
	FATAL severity = 0
	// ERROR is an option you can pass to Event to specify a severity of ERROR/1
	ERROR severity = 1
	// WARN is an option you can pass to Event to specify a severity of WARN/2
	WARN severity = 2
	// INFO is an option you can pass to Event to specify a severity of INFO/3
	INFO severity = 3
)

// severity is the log severity level
//
// we don't export this because we don't want the caller
// to define their own severity levels
type severity int

func (s severity) attach(le *EventData) {
	le.Severity = &s
}

// SeverityToLevel translates a log.go severity into a slog.Level
func SeverityToLevel(severity severity) slog.Level {
	switch severity {
	case FATAL:
		return LevelFatal
	case ERROR:
		return LevelError
	case WARN:
		return LevelWarn
	case INFO:
		return LevelInfo
	default:
		return LevelInfo
	}
}

// LevelToSeverity translates a slog.Level into a log.go severity
func LevelToSeverity(level slog.Level) severity {
	switch level {
	case LevelFatal:
		return FATAL
	case LevelError:
		return ERROR
	case LevelWarn:
		return WARN
	case LevelDebug, LevelInfo:
		return INFO
	default:
		return INFO
	}
}
