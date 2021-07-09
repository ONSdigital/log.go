log.go [![GoDoc](https://godoc.org/github.com/ONSdigital/log.go/log?status.svg)](https://godoc.org/github.com/ONSdigital/log.go/log)
======

A log library for Go.

Opinionated, and designed to match our [logging standards](https://github.com/ONSdigital/dp/blob/master/standards/LOGGING_STANDARDS.md).

### Getting started
Get the code:
```
git clone git@github.com:ONSdigital/log.go.git
```
**Note:** `log.go` is a Go Module so should be cloned outside your `$GOPATH`.

### Set up
To output logs in human readable format set the following environment var:
```bash
HUMAN_LOG=1
```

:warning: **This is for local dev use only** - DP developers should not enable human readable log output for apps running 
in an environment.

### Logging events
We recommend the first thing your `main` func does is to set the log `namespace`. Doing so will ensure that all log
events will be indexed correctly by Kibana. By convention the namespace should be the full repo name i.e. `dp-dataset-api`

Set the namespace:
```go
// set the log namespace
log.Namespace = "dp-logging-example"
```

Logging an INFO event example:
```go
// log an info event
log.Info(context.Background(), "info message with no additional data")
```
```json
{
  "created_at": "2020-12-10T11:16:39.155843Z",
  "event": "info message with no additional data",
  "namespace": "dp-logging-example",
  "severity": 3
}
```
Logging an INFO event with additional parameters example:
```go
// log an info event with additional parameters
log.Info(context.Background(), "info message with additional data", log.Data{
    "parma1": "value1",
    "parma2": "value2",
    "parma3": "value3",
})
```

```json
{
  "created_at": "2020-12-10T11:16:39.156147Z",
  "data": {
    "additional_data1": "value1",
    "additional_data2": "value2",
    "additional_data3": "value3"
  },
  "event": "info message with additional data",
  "namespace": "dp-logging-example",
  "severity": 3
}
```
Logging an ERROR event example:
```go
// log an error event
log.Error(context.Background(), "unexpected error", err)
```
```json
{
  "created_at": "2020-12-10T11:16:39.156205Z",
  "errors": [
    {
      "message": "something went wrong",
      "stack_trace": [
        {
          "file": "/Users/dave/Development/go/ons/log.go/example/main.go",
          "function": "main.main",
          "line": 27
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/proc.go",
          "function": "runtime.main",
          "line": 204
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/asm_amd64.s",
          "function": "runtime.goexit",
          "line": 1374
        }
      ]
    }
  ],
  "event": "unexpected error",
  "namespace": "dp-logging-example",
  "severity": 1
}
```
Logging an ERROR event with additional parameters example:
```go
// log an error event with additional parameters
log.Error(context.Background(), "unexpected error", err, log.Data{
    "additional_data": "some value",
})
```
```json
{
  "created_at": "2020-12-10T11:16:39.1564Z",
  "data": {
    "additional_data": "some value"
  },
  "errors": [
    {
      "message": "something went wrong",
      "stack_trace": [
        {
          "file": "/Users/dave/Development/go/ons/log.go/example/main.go",
          "function": "main.main",
          "line": 29
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/proc.go",
          "function": "runtime.main",
          "line": 204
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/asm_amd64.s",
          "function": "runtime.goexit",
          "line": 1374
        }
      ]
    }
  ],
  "event": "unexpected error",
  "namespace": "dp-logging-example",
  "severity": 1
}
```
Logging a custom error event that contains additional error data:
```go
// create a custom error
customError := &log.CustomError{
  Message: "something went wrong",
  Data:    map[string]interface{}{
    "error_code": 1093,
    "backing_service": "kafka",
    },
}

// log an error event where custom error has additional parameters (e.g. Data)
log.Error(context.Background(), "unexpected error", customErr)
```
```json
{
  "created_at": "2020-12-10T11:16:39.1564Z",
  "errors": [
    {
      "message": "something went wrong",
      "data": {
        "error_code": 1093,
        "backing_service": "kafka",
      },
      "stack_trace": [
        {
          "file": "/Users/dave/Development/go/ons/log.go/example/main.go",
          "function": "main.main",
          "line": 29
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/proc.go",
          "function": "runtime.main",
          "line": 204
        },
        {
          "file": "/usr/local/Cellar/go/1.15.2/libexec/src/runtime/asm_amd64.s",
          "function": "runtime.goexit",
          "line": 1374
        }
      ]
    }
  ],
  "event": "unexpected error",
  "namespace": "dp-logging-example",
  "severity": 1
}
```
Full code example:
```go
package main

import (
	"context"
	"errors"

	"github.com/ONSdigital/log.go/log"
)

func main() {
	// set the log namespace
	log.Namespace = "dp-logging-example"

	// log an info event
	log.Info(context.Background(), "info message with no additional data")

	// log an info event with additional parameters
	log.Info(context.Background(), "info message with additional data", log.Data{
		"parma1": "value1",
		"parma2": "value2",
		"parma3": "value3",
	})

	// an example error
	err := errors.New("something went wrong")

	// log an error event
	log.Error(context.Background(), "unexpected error", err)

	// log an error event with additional parameters
	log.Error(context.Background(), "unexpected error", err, log.Data{
		"additional_data": "some value",
	})

  // an example custom error
  customError := &log.CustomError{
    Message: "something went wrong",
    Data:    map[string]interface{}{
      "error_code": 1093,
      "backing_service": "kafka",
      },
  }

  // Log an error event with additional parameters
  log.Error(context.Background(), "unexpected error", customErr)
}
```
**Notes:**

- `context` can be nil however it's recommended to provide a `ctx` value if you have it available - internally 
  `log.<event e.g. Info, Warn, Error, Fatal>()` will automatically extract certain common fields (e.g. request IDs, http details) if they exist and add 
  them to the `log.Data` parameters map - this helps to ensure events contain as much useful information as possible.
  

- The `event` string should be a generic consistent message e.g. `http request received`. It should not format 
  additional values - these should be added to `log.Data` (see logging standards doc for a comprehensive overview).
  

- The `log.Event()` interface does not require you to provide a log (severity) level but it's recommended you provide this 
  field if possible/where appropriate. Better yet use the Wrapper functions `log.Info(...)`, `log.Warn(...)`, `log.Error(...)` and `log.Fatal(...)` to inherit log level.

### Scripts

* [edit-logs.sh](scripts) - helpful script to assist the updating of go-ns logs to v1 log.go logs package; it covers the majority of old logging styles from go-ns and converts them into expected logs that are compatible with version 1 of this library.

### Licence

Copyright ©‎ 2019-2021, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
