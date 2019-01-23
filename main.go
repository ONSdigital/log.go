package main

import (
	"net/http"
	"time"

	"github.com/ONSdigital/log.go/log"
)

var req, _ = http.NewRequest("GET", "/", nil)
var start, end = time.Now(), time.Now()

func main() {
	log.Event("started app")

	log.Event("received request", log.HTTP(req, 200, start, end))

	log.Event("doing something", log.Data{"key": "value"})

	log.Event("doing something", log.Data{"key": "value"}, log.HTTP(req, 401, start, end))
}
