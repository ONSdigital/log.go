package scripts

import (
	"context"
	"errors"

	// "github.com/ONSdigital/log.go/v2/log"
	"github.com/ONSdigital/go-ns/log"
)

func main() {
	// Setup test data for log messages
	ctx := context.Background()
	err := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"

	logData := log.Data{}

	// Once script has run against this file the error logs should update to the
	// comment above each one with the exception that err is reffered to as error

	// log.Event(ctx, , log.Message(error))
	log.Error(err, nil)

	// log.Event(ctx, , log.Message(error), logData)
	log.Error(err, logData)

	// log.Event(ctx, , log.Message(error), log.Data{"data_1": data1})
	log.Error(err, log.Data{"data_1": data1})

	// log.Event(ctx, "message", log.Message(error), log.Data{"data_1": data1, "data_2": data2})
	log.ErrorC("message", err, log.Data{"data_1": data1, "data_2": data2})

	// log.Event(ctx, message1, log.Message(error), log.Data{"data_1": data1, "data_2": data2})
	log.ErrorC(message1, err, log.Data{"data_1": data1, "data_2": data2})

	// log.Event(ctx, , log.Message(error), log.Data{"data_1": data1})
	log.ErrorCtx(ctx, err, log.Data{"data_1": data1})

	// log.Event(ctx, "message", )
	log.Info("message", nil)

	// log.Event(ctx, "message", log.Message(error), log.Data{"data-1": data1, "data-2": data2})
	log.InfoCtx(ctx, "message", log.Data{"error": err, "data-1": data1, "data-2": data2})

	// log.Event(ctx, "message", log.Message(error), log.Data{"data-1": data1, "data-2": data2, })
	log.InfoCtx(ctx, "message", log.Data{"data-1": data1, "data-2": data2, "error": err})

	// log.Event(ctx, "message", log.Message(error), log.Data{"data-1": data1, "data-2": data2, "data-3": data3, "data-4": data4})
	log.InfoCtx(ctx, "message", log.Data{"data-1": data1, "data-2": data2, "error": err, "data-3": data3, "data-4": data4})

	// log.Event(ctx, "message", log.Data{"data-1": data1, "data-2": data2, "data-3": data3})
	log.InfoCtx(ctx, "message", log.Data{"data-1": data1, "data-2": data2, "data-3": data3})
}
