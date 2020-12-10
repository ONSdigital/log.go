log.go [![Build Status](https://travis-ci.org/ONSdigital/log.go.svg?branch=master)](https://travis-ci.org/ONSdigital/log.go) [![GoDoc](https://godoc.org/github.com/ONSdigital/log.go/log?status.svg)](https://godoc.org/github.com/ONSdigital/log.go/log)
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
// Set the log namespace
log.Namespace = "dp-logging-example"
```

Logging an INFO event example:
```go
// Log an INFO event
log.Event(context.Background(), "info message with no additional data", log.INFO)
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
// Log an INFO event with additional parameters
log.Event(context.Background(), "info message with additional data", log.INFO, log.Data{
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
// Log an ERROR event
log.Event(context.Background(), "unexpected error", log.ERROR, log.Error(err))
```
```json
{
  "created_at": "2020-12-10T11:16:39.156205Z",
  "error": {
    "data": {},
    "error": "something went wrong",
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
  },
  "event": "unexpected error",
  "namespace": "dp-logging-example",
  "severity": 1
}
```
Logging an ERROR event with additional parameters example:
```go
// Log an ERROR event with additional parameters
log.Event(context.Background(), "unexpected error", log.ERROR, log.Error(err), log.Data{
    "additional_data": "some value",
})
```
```json
{
  "created_at": "2020-12-10T11:16:39.1564Z",
  "data": {
    "additional_data": "some value"
  },
  "error": {
    "data": {},
    "error": "something went wrong",
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
  },
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
	// Set the log namespace
	log.Namespace = "dp-logging-example"

	// Log an INFO event
	log.Event(context.Background(), "info message with no additional data", log.INFO)

	// Log an INFO event with additional parameters
	log.Event(context.Background(), "info message with additional data", log.INFO, log.Data{
		"parma1": "value1",
		"parma2": "value2",
		"parma3": "value3",
	})

	// an example error
	err := errors.New("something went wrong")

	// Log an ERROR event
	log.Event(context.Background(), "unexpected error", log.ERROR, log.Error(err))

	// Log an ERROR event with additional parameters
	log.Event(context.Background(), "unexpected error", log.ERROR, log.Error(err), log.Data{
		"additional_data": "some value",
	})
}
```
**Notes:**

- `context` can be nil however it's recommended to provide a `ctx` value if you have it available - internally 
  `log.Event()` will automatically extract certain common fields (e.g. request IDs, http details) if they exist and add 
  them to the `log.Data` parameters map - this helps to ensure events contain as much useful information as possible.
  

- The `event` string should be a generic consistent message e.g. `http request received`. It should not format 
  additional values - these should be added to `log.Data` (see logging standards doc for a comprehensive overview).
  

- The `log.Event()` interface does not require you to provide an event level but it's recommended you provide this 
  field if possible/where appropriate.

### Scripts

* [edit-logs.sh](scripts) - helpful script to assist the updating of go-ns logs to log.go logs; it covers the majority of old logging styles from go-ns and converts them into expected logs that are compatible with this library.

### Licence

Copyright ©‎ 2019-2020, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
