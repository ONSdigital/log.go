package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ONSdigital/log.go/log"
)

var req, _ = http.NewRequest("GET", "/", nil)
var start, end = time.Now(), time.Now()
var ctx = context.TODO()

func main() {
	log.Namespace = "log-test"

	log.Event(ctx, "started app")

	log.Event(ctx, "received request", log.HTTP(req, 200, &start, &end))

	log.Event(ctx, "doing something", log.Data{"key": "value"})

	log.Event(ctx, "doing something", log.Data{"key": "value"}, log.HTTP(req, 401, &start, &end))

	log.Event(ctx, "doing something", log.FATAL)

	log.Event(ctx, "doing something", log.FATAL, log.INFO, log.ERROR)

	log.Event(ctx, "doing something", log.Auth(log.USER, "user-id"))
	log.Event(ctx, "doing something", log.Auth(log.SERVICE, "service-id"))

	go http.ListenAndServe(":10203", log.Middleware(http.HandlerFunc(handler)))

	time.Sleep(5 * time.Millisecond)
	http.Get("http://localhost:10203")

	causeError()
}

func handler(w http.ResponseWriter, req *http.Request) {
	log.Event(req.Context(), "doing something in a handler")
}

func causeError() {
	log.Event(nil, "some error", log.Error(errors.New("foo")))
}

// func ctx() {
// 	// first arg
// 	log.Event(nil, "started app")
// 	log.Event(ctx, "started app")

// 	// optional arg
// 	log.Event("started app", log.Context(ctx))

// 	log.
// }
