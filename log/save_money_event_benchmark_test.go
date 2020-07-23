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

func BenchmarkLogSave(b *testing.B) {
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
	// Using the THREE new events also drops bytes allocated from ~ 850 to 769
	// (before dropping items)
	for i := 0; i < b.N; i++ {
		// 1st Event is like the first one in Middleware()
		SaveMoneyEvent1(ctx, "http request received", HTTP(req, 0, 0, &start, nil))

		// 2nd event is 'similar in length' to one in createReverseProxy()
		// BUT with items dropped, see file: log_line_length_Optimization-2.odt
		SaveMoneyEvent2(ctx, "proxying request", INFO, HTTP(req2, 0, 0, nil, nil),
			Data{"destination": babbageURL,
				"proxy_name": "babbage"})

		// 3rd Event is like the second one in Middleware()
		// BUT with items dropped, see file: log_line_length_Optimization-2.odt
		SaveMoneyEvent3(ctx, "http request completed", HTTP(req2, 200, 4, nil, &end))
	}
}
