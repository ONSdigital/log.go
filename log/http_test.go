package log

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTP(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:1234/a/b/c?x=1&y=2", nil)

	Convey("HTTP function returns a *EventHTTP", t, func() {
		eventHTTP := HTTP(req, 0, 0, nil, nil)
		So(eventHTTP, ShouldHaveSameTypeAs, &EventHTTP{})
		So(eventHTTP, ShouldImplement, (*option)(nil))

		Convey("*EventHTTP has the correct fields", func() {
			startTime := time.Now().UTC().Add(time.Second * -1)
			endTime := time.Now().UTC()
			duration := endTime.Sub(startTime)

			eventHTTP := HTTP(req, 101, 123, &startTime, &endTime)
			httpEvent := eventHTTP.(*EventHTTP)

			So(httpEvent.StatusCode, ShouldNotBeNil)
			So(*httpEvent.StatusCode, ShouldEqual, 101)
			So(httpEvent.Method, ShouldEqual, "GET")

			So(httpEvent.Scheme, ShouldEqual, "http")
			So(httpEvent.Host, ShouldEqual, "localhost")
			So(httpEvent.Port, ShouldEqual, 1234)
			So(httpEvent.Path, ShouldEqual, "/a/b/c")
			So(httpEvent.Query, ShouldEqual, "x=1&y=2")

			So(httpEvent.StartedAt, ShouldEqual, &startTime)
			So(httpEvent.EndedAt, ShouldEqual, &endTime)
			So(httpEvent.Duration, ShouldNotBeNil)
			So(*httpEvent.Duration, ShouldEqual, duration)
			So(httpEvent.ResponseContentLength, ShouldEqual, 123)
		})
	})

	Convey("*EventHTTP can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.HTTP, ShouldBeNil)

		eventHTTP := EventHTTP{}
		eventHTTP.attach(event)

		So(event.HTTP, ShouldResemble, &eventHTTP)
	})

	Convey("Duration should be nil if startedAt is nil", t, func() {
		endTime := time.Now().UTC()
		eventHTTP := HTTP(req, 101, 123, nil, &endTime)
		httpEvent := eventHTTP.(*EventHTTP)

		So(httpEvent.Duration, ShouldBeNil)
	})

	Convey("Duration should be nil if endedAt is nil", t, func() {
		startTime := time.Now().UTC()
		eventHTTP := HTTP(req, 101, 123, &startTime, nil)
		httpEvent := eventHTTP.(*EventHTTP)

		So(httpEvent.Duration, ShouldBeNil)
	})

}
