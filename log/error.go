package log

import (
	"reflect"
	"runtime"
)

// EventError is the data structure used for logging a error event.
//
// It isn't very useful to export, other than for documenting the
// data structure it outputs.
type EventError struct {
	Message    string            `json:"message,omitempty"`
	StackTrace []EventStackTrace `json:"stack_trace,omitempty"`
	// This uses interface{} type, but should always be a type of kind struct
	// (which serialises to map[string]interface{})
	// See `func Error` switch block for more info
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

type EventErrors []EventError

func (l *EventErrors) attach(le *EventData) {
	le.Errors = l
}

// FormatError returns an option you can pass to Event to attach
// error information to a log event
//
// It uses error.Error() to stringify the error value
//
// It also includes the error type itself as unstructured log
// data. For a struct{} type, it is included directly. For all
// other types, it is wrapped in a Data{} struct
//
// It also includes a full strack trace to where FormatError() is called,
// so you shouldn't normally store a log.Error for reuse (e.g. as a
// package level variable)
func FormatErrors(err []error) option {

	var e []EventError

	for i := range err {

		errs := EventError{
			Message:    err[i].Error(),
			StackTrace: make([]EventStackTrace, 0),
		}

		k := reflect.Indirect(reflect.ValueOf(err[i])).Type().Kind()
		switch k {
		case reflect.Struct:
			// We've got a struct type, so make it the top level value
			errs.Data = err[i]
		default:
			// We have something else, so nest it inside a Data value
			errs.Data = Data{"value": err[i]}
		}

		pc := make([]uintptr, 10)
		n := runtime.Callers(2, pc)
		if n > 0 {
			frames := runtime.CallersFrames(pc[:n])

			for {
				frame, more := frames.Next()

				errs.StackTrace = append(errs.StackTrace, EventStackTrace{
					File:     frame.File,
					Line:     frame.Line,
					Function: frame.Function,
				})

				if !more {
					break
				}
			}
		}

		e = append(e, errs)
	}

	a := EventErrors(e)

	return &a
}
