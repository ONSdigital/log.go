package main

import "testing"

func TestMain(t *testing.T) {
	log.Event(nil, "test", log.Data{}, log.Data{})
}
