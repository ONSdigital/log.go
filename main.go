package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/log.go/log"
)

func main() {
	fmt.Println("Check, before: Benchmarking: 'Log'")
	ctx := context.Background()
	errToLog := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"
	req, err := http.NewRequest("GET", "https://httpbin.org/get", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Event(ctx,
		message1,
		log.INFO,
		log.Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4},
		log.Error(errToLog),
		log.HTTP(req, 0, 0, nil, nil),
		log.Auth(log.USER, "tester-1"))
}
