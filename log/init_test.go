package log

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/ONSdigital/log.go/v3/config"

	. "github.com/smartystreets/goconvey/convey"
)

const someMessage = "some message goes here"
const someErrorMessage = "some error message"

var (
	someTimeLocal = time.Date(2024, time.June, 21, 14, 44, 50, 17, time.FixedZone("UTC+1", 1*60*60))
	someTimeUTC   = time.Date(2024, time.June, 21, 13, 44, 50, 17, time.UTC)
)

func TestLogger(t *testing.T) {
	Convey("With an empty logger var", t, func() {
		var handler slog.Handler

		Convey("Calling Handler with no options returns a logger", func() {
			handler = Handler()
			So(handler, ShouldNotBeNil)
		})

		Convey("Calling Handler with pretty returns a logger", func() {
			handler = Handler(config.Pretty)
			So(handler, ShouldNotBeNil)
		})

		Convey("Calling Handler with a namespace returns a logger", func() {
			handler = Handler(config.Namespace("some new namespace"))
			So(handler, ShouldNotBeNil)
		})
	})
}

func Test_replaceAttr(t *testing.T) {
	Convey("With a range of test cases", t, func() {
		tests := []struct {
			name   string
			groups []string
			attr   slog.Attr
			wanted slog.Attr
		}{
			{`msg becomes event`,
				nil,
				slog.Attr{
					Key:   "msg",
					Value: slog.StringValue(someMessage),
				},
				slog.Attr{
					Key:   "event",
					Value: slog.StringValue(someMessage),
				},
			},
			{"time becomes created_at",
				nil,
				slog.Attr{
					Key:   "time",
					Value: slog.TimeValue(someTimeLocal),
				},
				slog.Attr{
					Key:   "created_at",
					Value: slog.TimeValue(someTimeUTC),
				},
			},
			{`fatal level becomes "FATAL"`,
				nil,
				slog.Attr{
					Key:   "level",
					Value: slog.AnyValue(LevelFatal),
				},
				slog.Attr{
					Key:   "level",
					Value: slog.StringValue("FATAL"),
				},
			},
			{`error becomes an error struct`,
				nil,
				slog.Attr{
					Key:   "error",
					Value: slog.AnyValue(errors.New(someErrorMessage)),
				},
				slog.Attr{
					Key: "errors",
					Value: slog.AnyValue([]EventError{
						{
							Message:    "some error message",
							StackTrace: []EventStackTrace{},
						},
					}),
				},
			},
		}

		for _, tc := range tests {
			Convey(tc.name, func() {
				result := replaceAttr(tc.groups, tc.attr)
				So(result, ShouldResemble, tc.wanted)
			})
		}
	})
}
