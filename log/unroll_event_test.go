package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogNewUnrollFuncs(t *testing.T) {
	buf := &bytes.Buffer{}
	Convey("unrollIntToBuf2", t, func() {
		type nums struct {
			Number int
			Text   string
		}

		array := []nums{
			{0, "00"}, {1, "01"}, {2, "02"}, {3, "03"}, {4, "04"},
			{5, "05"}, {6, "06"}, {7, "07"}, {8, "08"}, {9, "09"},
			{10, "10"}, {11, "11"}, {23, "23"}, {34, "34"}, {45, "45"},
			{56, "56"}, {67, "67"}, {78, "78"}, {89, "89"}, {99, "99"},
			{100, "00"},
		}
		for _, v := range array {
			unrollIntToBuf2(buf, v.Number)
			So(buf.String(), ShouldEqual, v.Text)
			buf.Reset()
		}
	})

	Convey("unrollIntToBuf4", t, func() {
		type nums struct {
			Number int
			Text   string
		}

		array := []nums{
			{0, "0000"}, {1, "0001"}, {2, "0002"}, {3, "0003"}, {4, "0004"},
			{5, "0005"}, {6, "0006"}, {7, "0007"}, {8, "0008"}, {9, "0009"},
			{10, "0010"}, {11, "0011"}, {23, "0023"}, {34, "0034"}, {45, "0045"},
			{56, "0056"}, {67, "0067"}, {78, "0078"}, {89, "0089"}, {99, "0099"},
			{100, "0100"}, {203, "0203"}, {1100, "1100"}, {8765, "8765"}, {9999, "9999"},
			{10000, "0000"},
		}
		for _, v := range array {
			unrollIntToBuf4(buf, v.Number)
			So(buf.String(), ShouldEqual, v.Text)
			buf.Reset()
		}
	})

	Convey("unrollIntToBuf9", t, func() {
		type nums struct {
			Number int
			Text   string
		}

		array := []nums{
			{123456789, "123456789Z"},
			{123456780, "12345678Z"},
			{123456700, "1234567Z"},
			{123456000, "123456Z"},
			{123450000, "12345Z"},
			{123400000, "1234Z"},
			{123000000, "123Z"},
			{120000000, "12Z"},
			{100000000, "1Z"},
			{000000000, "0Z"},
			{000000001, "000000001Z"},
		}
		buf.Reset()
		for _, v := range array {
			unrollIntToBuf9(buf, v.Number)
			So(buf.String(), ShouldEqual, v.Text)
			buf.Reset()
		}
	})

	Convey("unrollInt", t, func() {
		type nums struct {
			Number int
			Text   string
		}

		array := []nums{
			{0, "0"}, {1, "1"}, {2, "2"}, {3, "3"}, {4, "4"},
			{5, "5"}, {6, "6"}, {7, "7"}, {8, "8"}, {9, "9"},
			{10, "10"}, {11, "11"}, {23, "23"}, {34, "34"}, {45, "45"},
			{56, "56"}, {67, "67"}, {78, "78"}, {89, "89"}, {99, "99"},
			{100, "100"}, {-1, "-1"}, {-9923, "-9923"},
			{1234567890, "1234567890"},
		}
		for _, v := range array {
			unrollInt(buf, v.Number)
			So(buf.String(), ShouldEqual, v.Text)
			buf.Reset()
		}
	})

	Convey("unrollInt64", t, func() {
		type nums struct {
			Number int64
			Text   string
		}

		array := []nums{
			{0, "0"}, {1, "1"}, {2, "2"}, {3, "3"}, {4, "4"},
			{5, "5"}, {6, "6"}, {7, "7"}, {8, "8"}, {9, "9"},
			{10, "10"}, {11, "11"}, {23, "23"}, {34, "34"}, {45, "45"},
			{56, "56"}, {67, "67"}, {78, "78"}, {89, "89"}, {99, "99"},
			{100, "100"}, {-1, "-1"}, {-9923, "-9923"},
			{1234567890, "1234567890"}, {1234567890123456789, "1234567890123456789"},
		}
		for _, v := range array {
			unrollInt64(buf, v.Number)
			So(buf.String(), ShouldEqual, v.Text)
			buf.Reset()
		}
	})

}

func TestLogNew3eventsRouter(t *testing.T) {
	// Test 3 events that look like what dp-frontend-router issues on the HAPPY HOT-PATH
	// Get the old events for the 3 and the new events for 3 and compare ...
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
	oldBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer1 = append(oldBuffer1, bytesWritten[i])
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	oldBuffer2 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer2 = append(oldBuffer2, bytesWritten[i])
	}

	// 3rd Event is like the second one in Middleware()
	Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))

	oldBuffer3 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer3 = append(oldBuffer3, bytesWritten[i])
	}

	//////////////////////
	// Capture NEW events

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	// 1st Event is like the first one in Middleware()
	Event(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

	newBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer1 = append(newBuffer1, bytesWritten[i])
	}

	// 2nd event is 'similar in length' to one in createReverseProxy()
	Event(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
		Data{"destination": babbageURL,
			"proxy_name": "babbage"})

	newBuffer2 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer2 = append(newBuffer2, bytesWritten[i])
	}

	// 3rd Event is like the second one in Middleware()
	Event(ctx, "http request completed", HTTP(req2, 200, 4, &start, &end))

	newBuffer3 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer3 = append(newBuffer3, bytesWritten[i])
	}

	//////////////////////
	// Compare old to new

	// The only difference should be in the 'created_at' timestamps

	if err := compareEvents("Event 1", oldBuffer1, newBuffer1); err != nil {
		o1 := bytes.NewBuffer(oldBuffer1)
		fmt.Fprintln(oldDestination, "Captured Event OLD 1:")
		o1.WriteTo(oldDestination) // ignore any error from this as it is not important

		n1 := bytes.NewBuffer(newBuffer1)
		fmt.Fprintln(oldDestination, "Captured Event NEW 1:")
		n1.WriteTo(oldDestination)

		t.Errorf("%v", err)
	}

	if err := compareEvents("Event 2", oldBuffer2, newBuffer2); err != nil {
		o2 := bytes.NewBuffer(oldBuffer2)
		fmt.Fprintln(oldDestination, "Captured Event OLD 2:")
		o2.WriteTo(oldDestination)

		n2 := bytes.NewBuffer(newBuffer2)
		fmt.Fprintln(oldDestination, "Captured Event NEW 2:")
		n2.WriteTo(oldDestination)

		t.Errorf("%v", err)
	}

	if err := compareEvents("Event 3", oldBuffer3, newBuffer3); err != nil {
		o3 := bytes.NewBuffer(oldBuffer3)
		fmt.Fprintln(oldDestination, "Captured Event OLD 3:")
		o3.WriteTo(oldDestination)

		n3 := bytes.NewBuffer(newBuffer3)
		fmt.Fprintln(oldDestination, "Captured Event NEW 3:")
		n3.WriteTo(oldDestination)

		t.Errorf("%v", err)
	}
}

func compareEvents(eventNumber string, eventOld []byte, eventNew []byte) error {
	//eventOld = append(eventOld, byte(10)) // add termination for json Unmarshal to know when to stop
	//eventNew = append(eventNew, byte(10)) // add termination for json Unmarshal to know when to stop

	var to1, tn1 EventData2

	err := json.Unmarshal(eventOld, &to1)
	if err != nil {
		fmt.Printf("Problem with Old %s\n", eventNumber)
		return err
	}
	err = json.Unmarshal(eventNew, &tn1)
	if err != nil {
		fmt.Printf("Problem with New %s\n", eventNumber)
		return err
	}

	if to1.CreatedAt == tn1.CreatedAt {
		es := eventNumber + ": 'created_at' should not be the same"
		return errors.New(es)
	}

	if to1.Namespace != tn1.Namespace {
		es := eventNumber + ": 'namespace' should be the same"
		return errors.New(es)
	}

	if to1.Event != tn1.Event {
		es := eventNumber + ": 'event' should be the same"
		return errors.New(es)
	}

	if to1.TraceID != tn1.TraceID {
		es := eventNumber + ": 'trace_id' should be the same"
		return errors.New(es)
	}

	if to1.Severity != nil && tn1.Severity != nil {
		if *to1.Severity != *tn1.Severity {
			es := eventNumber + ": 'severity' should be the same"
			return errors.New(es)
		}
	}

	if to1.HTTP != nil && tn1.HTTP != nil {
		if to1.HTTP.StatusCode != nil && tn1.HTTP.StatusCode != nil {
			if *to1.HTTP.StatusCode != *tn1.HTTP.StatusCode {
				es := eventNumber + ": 'status_code' should be the same"
				return errors.New(es)
			}
		}

		if to1.HTTP.Method != tn1.HTTP.Method {
			es := eventNumber + ": 'method' should be the same"
			return errors.New(es)
		}
		if to1.HTTP.Scheme != tn1.HTTP.Scheme {
			es := eventNumber + ": 'scheme' should be the same"
			return errors.New(es)
		}
		if to1.HTTP.Host != tn1.HTTP.Host {
			es := eventNumber + ": 'host' should be the same"
			return errors.New(es)
		}
		if to1.HTTP.Port != tn1.HTTP.Port {
			es := eventNumber + ": 'port' should be the same"
			return errors.New(es)
		}
		if to1.HTTP.Path != tn1.HTTP.Path {
			es := eventNumber + ": 'path' should be the same"
			return errors.New(es)
		}
		if to1.HTTP.Query != tn1.HTTP.Query {
			es := eventNumber + ": 'query' should be the same"
			return errors.New(es)
		}

		if to1.HTTP.StartedAt != nil && tn1.HTTP.StartedAt != nil {
			if *to1.HTTP.StartedAt != *tn1.HTTP.StartedAt {
				es := eventNumber + ": 'started_at' should be the same"
				return errors.New(es)
			}
		}

		if to1.HTTP.EndedAt != nil && tn1.HTTP.EndedAt != nil {
			if *to1.HTTP.EndedAt != *tn1.HTTP.EndedAt {
				es := eventNumber + ": 'ended_at' should be the same"
				return errors.New(es)
			}
		}

		if to1.HTTP.Duration != nil && tn1.HTTP.Duration != nil {
			if *to1.HTTP.Duration != *tn1.HTTP.Duration {
				es := eventNumber + ": 'duration' should be the same"
				return errors.New(es)
			}
		}

		if to1.HTTP.ResponseContentLength != tn1.HTTP.ResponseContentLength {
			es := eventNumber + ": 'response_content_length' should be the same"
			return errors.New(es)
		}
	}

	if to1.Auth != nil && tn1.Auth != nil {
		//tn1.Auth.Identity = "" // put this in to 'test' test code
		eq := reflect.DeepEqual(to1.Auth, tn1.Auth)
		if !eq {
			fmt.Printf("old: %+v\n", to1.Auth)
			fmt.Printf("new: %+v\n", tn1.Auth)

			es := eventNumber + ": 'auth' should be the same"
			return errors.New(es)
		}
	}

	if to1.Data != nil && tn1.Data != nil {
		// Extract data to map[] for easy printing if there is a problem.
		oldData := make(map[string]string)

		for k, v := range *to1.Data {
			oldData[k] = fmt.Sprintf("%+v", v)
		}

		newData := make(map[string]string)

		for k, v := range *tn1.Data {
			newData[k] = fmt.Sprintf("%+v", v)
		}

		//newData["proxy_name"] = "test" // put this in to 'test' test code
		eq := reflect.DeepEqual(oldData, newData)
		if !eq {
			for k, v := range oldData {
				fmt.Printf("old: %s, %+v\n", k, v)
			}
			for k, v := range newData {
				fmt.Printf("new: %s, %+v\n", k, v)
			}

			es := eventNumber + ": 'data' should be the same"
			return errors.New(es)
		}
	}

	if to1.Error != nil && tn1.Error != nil {
		//tn1.Error.Error = "" // put this in to 'test' test code
		eq := reflect.DeepEqual(to1.Error, tn1.Error)
		if !eq {
			fmt.Printf("old: %+v\n", to1.Error)
			fmt.Printf("new: %+v\n", tn1.Error)

			es := eventNumber + ": 'error' should be the same"
			return errors.New(es)
		}
	}

	return nil
}

func TestLogNew1eventAll(t *testing.T) {
	// Test 1 event with all options passed.
	// Get the old event1 and the new event and compare ...

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

	fmt.Println("Testing: 'New Log 1 All options")
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

	errToLog := errors.New("test error")

	//	start := time.Now().UTC()
	//	end := time.Now().UTC()
	babbageURL, err := url.Parse("http://localhost:8080")

	Namespace = "BenchmarkLog" // force a fixed value as sometimes during testing this changes and does not help when comparing to other tests

	//////////////////////
	// Capture old event

	isMinimalAllocations = false // use existing Event() code

	eventError := Error(errToLog) // pull this out here to have same 'Line number' in stack trace

	// 1st Event is like the first one in Middleware()
	// Capture the output of the call to Event()
	var bytesWritten []byte
	destination = &writer{func(b []byte) (n int, err error) {
		bytesWritten = b
		return len(b), nil
	}}
	Event(ctx, "http request received",
		eventError,
		HTTP(req, 0, 0, nil, nil),
		Data{"destination": babbageURL, "proxy_name": "babbage"},
		Auth(USER, "tester-1"))

	// Converting what has been captured in bytesWritten with string()
	// puts : !F(MISSING)
	// into the output, so we do the following:
	// We have to copy the result into a new buffer because the Fprintln over-writes
	// the result (what a pain).
	oldBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		oldBuffer1 = append(oldBuffer1, bytesWritten[i])
	}

	//////////////////////
	// Capture NEW events

	isMinimalAllocations = true // use new Event() code, for minimum memory allocations

	// 1st Event is like the first one in Middleware()
	Event(ctx, "http request received",
		eventError,
		HTTP(req, 0, 0, nil, nil),
		Data{"destination": babbageURL, "proxy_name": "babbage"},
		Auth(USER, "tester-1"))

	newBuffer1 := make([]byte, 0, 2000)
	for i := 0; i < len(bytesWritten); i++ {
		newBuffer1 = append(newBuffer1, bytesWritten[i])
	}

	//////////////////////
	// Compare old to new

	// The only difference should be in the 'created_at' timestamps

	if err := compareEvents("Event 1", oldBuffer1, newBuffer1); err != nil {
		o1 := bytes.NewBuffer(oldBuffer1)
		fmt.Fprintln(oldDestination, "Captured Event OLD 1:")
		o1.WriteTo(oldDestination) // ignore any error from this as it is not important

		n1 := bytes.NewBuffer(newBuffer1)
		fmt.Fprintln(oldDestination, "Captured Event NEW 1:")
		n1.WriteTo(oldDestination)

		t.Errorf("%v", err)
	}
}

func TestLogNew3eventsRouterReduced(t *testing.T) {
	// !!! add comments + code
}
