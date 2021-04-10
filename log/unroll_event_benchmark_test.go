package log

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"
)

// run with:
// go test -run=log_test.go -bench=Log -benchtime=100x

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
var requestIDRandom = rand.New(rand.NewSource(99)) // seed with constant to get same sequence out output for every benchmar run
var randMutex sync.Mutex

// NewRequestID generates a random string of requested length
func newRequestID(size int) string {
	b := make([]rune, size)
	randMutex.Lock()
	for i := range b {
		b[i] = letters[requestIDRandom.Intn(len(letters))]
	}
	randMutex.Unlock()
	return string(b)
}

func BenchmarkLog1(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log'")
	errToLog := errors.New("test error")
	message1 := "Benchmark test"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("req: %v\n", req)

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = false // use existing Event() code

	b.ReportAllocs()

	// test all event types
	for i := 0; i < b.N; i++ {
		Event(ctx,
			message1,
			INFO,
			Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4},
			Error(errToLog),
			HTTP(req, 0, 0, nil, nil),
			Auth(USER, "tester-1"))
	}
}

// run with:
// go test -run=log_test.go -bench=. -benchtime=1000000000x

// on 1st May 2020 gave results:
/*

Benchmarking: 'Log - o.attach'
goos: linux
goarch: amd64
pkg: github.com/ONSdigital/log.go/log
BenchmarkLog2-12    	Benchmarking: 'Log - o.attach'
1000000000	       142 ns/op
Benchmarking: 'Log - switch'
BenchmarkLog3-12    	Benchmarking: 'Log - switch'
1000000000	       141 ns/op
PASS
ok  	github.com/ONSdigital/log.go/log	282.637s

*/

func BenchmarkLog2(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log - o.attach'")
	err := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"

	var opts [4]option

	opts[0] = INFO
	opts[1] = Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4}
	opts[2] = Error(err)
	opts[3] = Data{"data_4": data4, "data_2": data2}

	e := EventData{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     message1,
	}

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	for i := 0; i < b.N; i++ {
		// loop around each log option and call its attach method, which takes care
		// of the association with the EventData struct
		for _, o := range opts {
			// Using rare pattern : `thing.attach(toObject)`
			// this handles both cases where:
			// the receiver can be called either `dataThing.attach(...)` or `ptrToDataThing.attach(...)
			o.attach(&e)
		}
	}
}

func BenchmarkLog3(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log - switch'")
	err := errors.New("test error")
	message1 := "m1"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"

	var opts [4]option

	opts[0] = INFO
	opts[1] = Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4}
	opts[2] = Error(err)
	opts[3] = Data{"data_4": data4, "data_2": data2}

	e := EventData{
		CreatedAt: time.Now().UTC(),
		Namespace: Namespace,
		Event:     message1,
	}

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	for i := 0; i < b.N; i++ {
		// loop around each log option and attach each option
		// directly into EventData struct
		for _, o := range opts {
			// Doing typical pattern : `object.attach(thing)`
			switch v := o.(type) {
			case severity:
				e.Severity = &v
			case *severity: // added to match o.attach(e) code for completness (may never be used)
				e.Severity = v
			case Data:
				e.Data = &v
			case *Data: // added to match o.attach(e) code for completness (may never be used)
				e.Data = v
			case *EventHTTP:
				e.HTTP = v
			case *EventError:
				e.Error = v
			case *eventAuth:
				e.Auth = v
			default:
				fmt.Printf("option: %v, %v, %T", o, v, v)
				panic("unknown option")
			}
		}
	}
}

func BenchmarkLog4(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

		// 2nd event is 'similar in length' to one in createReverseProxy()
		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		// 3rd Event is like the second one in Middleware()
		Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))
	}
}

func BenchmarkLog5(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = false // use existing Event() code

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

		// 2nd event is 'similar in length' to one in createReverseProxy()
		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		// 3rd Event is like the second one in Middleware()
		Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))
	}
}

func BenchmarkLog6(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log'")
	errToLog := errors.New("test error")
	message1 := "Benchmark test"
	data1 := "d1"
	data2 := "d2"
	data3 := "d3"
	data4 := "d4"
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("req: %v\n", req)

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()

	// test all event types
	for i := 0; i < b.N; i++ {
		Event(ctx,
			message1,
			INFO,
			Data{"data_1": data1, "data_2": data2, "data_3": data3, "data_4": data4},
			Error(errToLog),
			HTTP(req, 0, 0, nil, nil),
			Auth(USER, "tester-1"))
	}
}

func BenchmarkLog7(b *testing.B) {
	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	fmt.Println("Benchmarking: 'Log'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is revere proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	q := req2.URL.Query()                                                                                                                                                                // Get a copy of the query values.
	q.Add(":uri", "embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi") // Add a new value to the set.
	req2.URL.RawQuery = q.Encode()                                                                                                                                                       // Encode and assign back to the original query.

	requestID := newRequestID(16)
	ctx := context.WithValue(context.Background(), common.RequestIdKey, requestID)
	start := time.Now().UTC()
	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	b.ReportAllocs()
	// The sequence of these 3 events is about worst case length that dp-frontend-router can do
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		var statusCode int = 0

		port := 0
		if p := req.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}

		var duration *time.Duration

		// inline the the setting up of the "EventHTTP" to save doing the HTTP(...)
		// thing as this escapes to the heap, whereas doing the following stays within
		// the stack of this calling function.
		e := EventHTTP{
			StatusCode: &statusCode,
			Method:     req.Method,

			Scheme: req.URL.Scheme,
			Host:   req.URL.Hostname(),
			Port:   port,
			Path:   req.URL.Path,
			Query:  req.URL.RawQuery,

			StartedAt:             &start,
			EndedAt:               nil,
			Duration:              duration,
			ResponseContentLength: 0,
		}

		//Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))
		Event(ctx, "http request received", &e)

		port = 0
		if p := req2.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}

		e.Method = req2.Method
		e.Scheme = req2.URL.Scheme
		e.Host = req2.URL.Hostname()
		e.Port = port
		e.Path = req2.URL.Path
		e.Query = req2.URL.RawQuery
		e.StartedAt = nil

		// 2nd event is 'similar in length' to one in createReverseProxy()
		//		Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Event(ctx, "proxying request", INFO, &e,
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		port = 0
		if p := req2.URL.Port(); len(p) > 0 {
			port, _ = strconv.Atoi(p)
		}
		port = 20000

		d := end.Sub(start)

		e.Port = port
		e.StartedAt = &start
		e.EndedAt = &end
		e.Duration = &d
		e.ResponseContentLength = 4
		statusCode = 200

		e.Method = req2.Method
		e.Scheme = req2.URL.Scheme
		e.Host = req2.URL.Hostname()
		e.Port = port
		e.Path = req2.URL.Path
		e.Query = req2.URL.RawQuery

		// 3rd Event is like the second one in Middleware()
		//		Event(req.Context(), "http request completed", HTTP(req2, 200, 4, &start, &end))
		Event(ctx, "http request completed", &e)
	}
}
