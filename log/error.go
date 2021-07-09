package log

import (
	"fmt"
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

type EventErrors []EventError

func (l *EventErrors) attach(le *EventData) {
	le.Errors = l
}

// FormatErrors returns an option you can pass to Event to attach
// error information to a log event
//
// It uses error.Error() to stringify the error value
//
// It also includes the error type itself as unstructured log
// data. For a struct{} type, it is included directly. For all
// other types, it is wrapped in a Data{} struct
//
// It also includes a full strack trace to where FormatErrors() is called,
// so you shouldn't normally store a log.Error for reuse (e.g. as a
// package level variable)
func FormatErrors(errs []error) option {

	var e []EventError

	for i := range errs {

		err := EventError{
			Message:    errs[i].Error(),
			StackTrace: make([]EventStackTrace, 0),
		}

		k := reflect.Indirect(reflect.ValueOf(errs[i])).Type().Kind()
		switch k {
		case reflect.Struct:

			// Check error types
			switch newErr := errs[i].(type) {
			case nil:
				fmt.Println("\nerror does not match any error types")
			case *CustomError: // matched CustomError type
				err.Data = newErr.ErrorData()
			case error:
				// catch everything else
			}

		default:
			// We have something else, so nest it inside a Data value
			err.Data = Data{"value": errs[i]}
		}

		pc := make([]uintptr, 10)
		n := runtime.Callers(2, pc)
		if n > 0 {
			frames := runtime.CallersFrames(pc[:n])

			for {
				frame, more := frames.Next()

				err.StackTrace = append(err.StackTrace, EventStackTrace{
					File:     frame.File,
					Line:     frame.Line,
					Function: frame.Function,
				})

				if !more {
					break
				}
			}
		}

		e = append(e, err)
	}

	a := EventErrors(e)

	return &a
}

type CustomError struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func (c *CustomError) Error() string {
	return c.Message
}

func (c CustomError) ErrorData() map[string]interface{} {
	return c.Data
}
