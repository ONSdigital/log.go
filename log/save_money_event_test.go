package log

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"
)

func TestLogSaveMoneyEventSavings(t *testing.T) {
	// Create 3 events that look like what dp-frontend-router issues on the HAPPY HOT-PATH
	// Get the Lengths for the 3 old events for the 3 new events and print savings.
	// The 3 old events are captured first and copied to buffers and then the
	// 3 new events are captured and copied to buffers to ensure no mistakes in
	// what one 'thinks' is in a particular buffer.

	oldNamespace := Namespace
	defer func() {
		Namespace = oldNamespace // this needed for other test functions
	}()

	isTestMode = false

	oldDestination := destination
	oldFallbackDestination := fallbackDestination

	defer func() {
		destination = oldDestination
		fallbackDestination = oldFallbackDestination
		isMinimalAllocations = false // put back to use old events, so as to not damage existing tests
	}()

	fmt.Println("Testing: 'New Log 3 Events in dp-frontend-router HAPPY HOT-PATH'")
	req, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		t.Errorf("%v", err)
	}

	// NOTE: The gorilla library function registerVars() in pat.go V1.0.1
	//       adds in the the resulting path that is reverse proxied to.
	// SO: The following replicates that so that this test more closely
	//     matches what is seen in dp-frontend-router.
	req2, err := http.NewRequest("GET", "http://localhost:20000/embed/visualisations/peoplepopulationandcommunity/populationandmigration/internationalmigration/qmis/shortterminternationalmigrationestimatesforlocalauthoritiesqmi", nil)
	if err != nil {
		t.Errorf("%v", err)
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

	var oldLength1, oldLength2, oldLength3 int

	//////////////////////
	// Capture old events

	isMinimalAllocations = false // use existing Event() code

	// 1st Event is like the first one in Middleware()
	// Capture the output of the call to Event()
	var bytesWritten []byte
	destination = &writer{func(b []byte) (n int, err error) {
		bytesWritten = b
		return len(b), nil
	}}
	Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

	// Converting what has been captured in bytesWritten with string()
	// puts : !F(MISSING)
	// into the output, so we do the following:
	// We have to copy the result into a new buffer because the Fprintln over-writes
	// the result (what a pain).
	oldLength1 = len(bytesWritten)
	oldBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < oldLength1; i++ {
		oldBuffer1 = append(oldBuffer1, bytesWritten[i])
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	oldLength2 = len(bytesWritten)
	oldBuffer2 := make([]byte, 0, 2000)
	for i := 0; i < oldLength2; i++ {
		oldBuffer2 = append(oldBuffer2, bytesWritten[i])
	}

	// 3rd Event is like the second one in Middleware()
	Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))

	oldLength3 = len(bytesWritten)
	oldBuffer3 := make([]byte, 0, 2000)
	for i := 0; i < oldLength3; i++ {
		oldBuffer3 = append(oldBuffer3, bytesWritten[i])
	}

	//////////////////////
	// Capture NEW events

	var newLength1, newLength2, newLength3 int
	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	// 1st Event is like the first one in Middleware()
	SaveMoneyEvent1(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

	newLength1 = len(bytesWritten)
	newBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < newLength1; i++ {
		newBuffer1 = append(newBuffer1, bytesWritten[i])
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	SaveMoneyEvent2(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	newLength2 = len(bytesWritten)
	newBuffer2 := make([]byte, 0, 2000)
	for i := 0; i < newLength2; i++ {
		newBuffer2 = append(newBuffer2, bytesWritten[i])
	}

	// 3rd Event is like the second one in Middleware()
	// NOTE: &start is not passed as part of the savings
	SaveMoneyEvent3(ctx, "http request completed", HTTP(req2, 200, 4, nil, &end))

	newLength3 = len(bytesWritten)
	newBuffer3 := make([]byte, 0, 2000)
	for i := 0; i < newLength3; i++ {
		newBuffer3 = append(newBuffer3, bytesWritten[i])
	}

	/////////////////////////////
	// Show what the savings are:

	fmt.Printf("Event 1 : old Length: %d, new Length: %d\n", oldLength1, newLength1)
	fmt.Printf("OLD 1:\n%v\n", string(oldBuffer1))
	fmt.Printf("NEW 1:\n%v\n", string(newBuffer1))
	fmt.Printf("Event 2 : old Length: %d, new Length: %d\n", oldLength2, newLength2)
	fmt.Printf("OLD 2:\n%v\n", string(oldBuffer2))
	fmt.Printf("NEW 2:\n%v\n", string(newBuffer2))
	fmt.Printf("Event 2 : old Length: %d, new Length: %d\n", oldLength3, newLength3)
	fmt.Printf("OLD 3:\n%v\n", string(oldBuffer3))
	fmt.Printf("NEW 3:\n%v\n", string(newBuffer3))

	oldTotal := oldLength1 + oldLength2 + oldLength3
	newTotal := newLength1 + newLength2 + newLength3
	fmt.Printf("Initial totals old: %d, new %d\n", oldTotal, newTotal)

	var extra string = "\"scheme\":\"http\",\"host\":\"localhost\",\"port\":20000"
	fmt.Printf("In these tests, unfortunately the following appears in the old events:\n%s\n", extra)
	fmt.Printf("  which may not be in the logs from dp-frontend-router.")
	fmt.Printf("So, removing the length of the 'extra' info from the old events ...\n")

	oldTotal = oldTotal - (3 * len(extra))
	fmt.Printf("FINAL totals old: %d, new %d\n", oldTotal, newTotal)

	savings := (1.0 - (float32(newTotal) / float32(oldTotal))) * 100.0
	fmt.Printf("\n  ****  Thats a saving of %0.1f%%  ****\n\n", savings)
}
