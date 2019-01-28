package log

import "runtime"

type eventError struct {
	Error string            `json:"error,omitempty"`
	Frame []eventErrorFrame `json:"stack,omitempty"`
}

type eventErrorFrame struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

func (l *eventError) attach(le *EventData) {
	le.Error = l
}

// Error ...
func Error(err error) option {
	// FIXME do we want to capture `err` somewhere?
	e := &eventError{
		Error: err.Error(),
		Frame: make([]eventErrorFrame, 0),
	}

	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	if n > 0 {
		frames := runtime.CallersFrames(pc[:n])

		for {
			frame, more := frames.Next()

			e.Frame = append(e.Frame, eventErrorFrame{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
			})

			if !more {
				break
			}
		}
	}

	return e
}
