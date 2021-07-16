package log

import (
	"context"
)

type eventFuncMock struct {
	capCtx        context.Context
	capEvent      string
	capOpts       []option
	hasBeenCalled bool
	severity      severity

	onEvent func(e eventFuncMock)
}

func (e *eventFuncMock) Event(ctx context.Context, event string, severity severity, opts ...option) {
	e.capCtx = ctx
	e.capEvent = event
	e.capOpts = opts
	e.hasBeenCalled = true
	e.severity = severity

	if e.onEvent != nil {
		e.onEvent(eventFuncMock{
			capCtx:        ctx,
			capEvent:      event,
			capOpts:       opts,
			hasBeenCalled: true,
			severity:      severity,
		})
	}
}
