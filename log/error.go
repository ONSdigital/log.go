package log

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EventErrors is an array of error events
type EventErrors []EventError

// EventError is the data structure used for logging a error event.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventError struct {
	Message    string            `json:"message,omitempty"`
	StackTrace []EventStackTrace `json:"stack_trace,omitempty"`
	// This uses interface{} type, but should always be a type of kind struct
	// (which serialises to map[string]interface{})
	// See `func FormatErrors` switch block for more info
	Data interface{} `json:"data,omitempty"`
}

// EventStackTrace is the data structure used for logging a stack trace.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventStackTrace struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

// FormatAsErrors takes an error and unwraps and wrapped errors and returns this as a slice. If any of the wrapped
// errors also contain a stack trace, this will also be extracted.
// Currently, stack traces provided by the following errors packages are supported...
// - golang.org/x/xerrors
// - github.com/pkg/errors
func FormatAsErrors(err error) []EventError {
	evtErrs := make([]EventError, 0, 4)

	for wrappedErr := err; wrappedErr != nil; wrappedErr = errors.Unwrap(wrappedErr) {
		evtErrs = append(evtErrs, EventError{
			Message:    wrappedErr.Error(),
			StackTrace: extractStacktrace(wrappedErr),
		})
	}

	return evtErrs
}

func extractStacktrace(err error) []EventStackTrace {
	st := make([]EventStackTrace, 0)

	pkg := reflect.Indirect(reflect.ValueOf(err)).Type().PkgPath()
	switch pkg {
	case "golang.org/x/xerrors":
		st = extractXErrStacktrace(err)
	case "github.com/pkg/errors":
		st = extractPkgErrStacktrace(err)
	}

	return st
}

func extractXErrStacktrace(xerr error) []EventStackTrace {
	errstr := fmt.Sprintf("%+v", xerr)
	lines := strings.Split(errstr, "\n")
	function := strings.Trim(lines[1], " ")
	caller := strings.Split(strings.Trim(lines[2], " "), ":")
	file := caller[0]
	lineNum, _ := strconv.Atoi(caller[1])

	return []EventStackTrace{{
		File:     file,
		Line:     lineNum,
		Function: function,
	}}
}

func extractPkgErrStacktrace(pkgerr error) []EventStackTrace {
	var st []EventStackTrace

	errstr := fmt.Sprintf("%+v", pkgerr)
	lines := strings.Split(errstr, "\n")

	ststart := len(lines)
	if ststart < 2 {
		return st
	}
	for i := ststart - 1; lines[i][0] == '\t'; i -= 2 {
		ststart = i - 1
	}
	stLines := lines[ststart:]

	for i := 0; i < len(stLines); i += 2 {
		function := strings.Trim(stLines[i], " ")
		caller := strings.Split(strings.Trim(stLines[i+1], "\t "), ":")
		file := caller[0]
		lineNum, _ := strconv.Atoi(caller[1])

		st = append(st, EventStackTrace{
			File:     file,
			Line:     lineNum,
			Function: function,
		})
	}
	return st
}

func (l *EventErrors) attach(le *EventData) {
	le.Errors = l
}
