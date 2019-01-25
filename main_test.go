package main

import (
	"testing"

	"github.com/ONSdigital/log.go/log"
)

func TestMain(t *testing.T) {
	log.Event(nil, "test", log.Data{}, log.Data{})
}
