package log

import (
	"context"
	"log"
	"strings"
)

//nolint:gochecknoinits // Needs to be refactored.
func init() {
	// Set the output for the default go logger
	log.SetOutput(&captureLogger{})
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

type captureLogger struct{}

func (c captureLogger) Write(b []byte) (n int, err error) {
	Event(context.Background(), "third party logs", INFO, Data{"raw": strings.TrimSpace(string(b))})
	return len(b), nil
}
