package log

import (
	"log"
	"strings"
)

//nolint:gochecknoinits // The init function is necessary for setting up default logging.
func init() {
	// Set the output for the default go logger
	log.SetOutput(&captureLogger{})
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

type captureLogger struct{}

func (c captureLogger) Write(b []byte) (n int, err error) {
	//nolint:staticcheck // Passing nil context here is intentional
	Event(nil, "third party logs", INFO, Data{"raw": strings.TrimSpace(string(b))})
	return len(b), nil
}
