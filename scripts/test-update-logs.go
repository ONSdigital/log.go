// +build skip

package scripts

import (
	"context"
	"errors"

	"github.com/ONSdigital/log.go/v2/log"
)

func main() {
	// Setup test data for log messages
	ctx := context.Background()
	err := errors.New("test error")

	logData := log.Data{"field-1": "value 1"}

	// Once script has run against this file the error logs should update to the
	// comment above each one with the exception that err is reffered to as error

	// log.Info(ctx, "test message")
	log.Event(ctx, "test message", log.INFO)

	// log.Info(ctx, "test message", log.Data{"field-1": "value 1", "field-2": "value 2"})
	log.Event(ctx, "test message", log.INFO, log.Data{"field-1": "value 1", "field-2": "value 2"})

	// log.Info(ctx, "test message", logData)
	log.Event(ctx, "test message", log.INFO, logData)

	// log.Error(ctx, "test mesage", log.FormatErrors([]error{err}))
	log.Event(ctx, "test message", log.ERROR, log.Error(err))

	// log.Error(ctx, "test mesage", log.FormatErrors([]error{err}), logData)
	log.Event(ctx, "test message", log.ERROR, log.Error(err), logData)

	// log.Warn(ctx, "test mesage", log.FormatErrors([]error{err}))
	log.Event(ctx, "test message", log.WARN, log.Error(err))

	// log.Warn(ctx, "test mesage", log.FormatErrors([]error{err}), logData)
	log.Event(ctx, "test message", log.WARN, log.Error(err), logData)
}
