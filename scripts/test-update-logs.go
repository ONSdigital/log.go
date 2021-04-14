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

	// Once script has run against this file the error logs should update to the
	// comment above each one with the exception that err is reffered to as error

	// log.Event(ctx, , log.Message(error))
	log.Error(err)

	// log.Event(ctx, , log.Message(error), logData)
	log.Error(err)

	// log.Event(ctx, , log.Message(error), log.Data{"data_1": data1})
	log.Error(err)

	// log.Event(ctx, "message", )
	log.Info(ctx, "message", nil)

}
