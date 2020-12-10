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
**Formatted output**

To output logs in human readable format set the following environment var:
```bash
HUMAN_LOG=1
```

:warning: **This is for local dev use only** - DP developers should not enable human readable log output for apps running 
in an environment.

**namespace**

We recommend the first thing your `main` func does is to set the log `namespace`. Doing so will ensure that all log 
events will be indexed correctly by Kibana. By convention the namespace should be the full repo name i.e. `dp-dataset-api`


### Logging events

```go
    // setting the namespace should be one of the first things done in main. 
    log.Namespace = "dp-logging-example"
```
	
	// Log an INFO message with 
    log.Event(context.Background(), "info message with additional data", log.INFO, log.Data{
        "additional_data1": "value1",
        "additional_data2": "value2",
        "additional_data3": "value3",
    })
```


### Scripts

* [edit-logs.sh](scripts) - helpful script to assist the updating of go-ns logs to log.go logs; it covers the majority of old logging styles from go-ns and converts them into expected logs that are compatible with this library.

### Licence

Copyright ©‎ 2019-2020, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
