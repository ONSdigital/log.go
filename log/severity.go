package log

const (
	// FATAL ...
	FATAL Severity = iota
	// ERROR ...
	ERROR
	// WARN ...
	WARN
	// INFO ...
	INFO
)

// Severity is the log severity level
type Severity int

// Attach ...
func (s Severity) Attach(le *EventData) {
	le.Severity = &s
}
