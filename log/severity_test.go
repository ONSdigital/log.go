package log

import (
	"fmt"
	"log/slog"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSeverity(t *testing.T) {
	Convey("severity can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Severity, ShouldBeNil)

		FATAL.attach(event)
		So(event.Severity, ShouldNotBeNil)
		So(*event.Severity, ShouldEqual, FATAL)
	})

	Convey("severity values are of type severity", t, func() {
		So(FATAL, ShouldHaveSameTypeAs, severity(-1))
		So(ERROR, ShouldHaveSameTypeAs, severity(-1))
		So(WARN, ShouldHaveSameTypeAs, severity(-1))
		So(INFO, ShouldHaveSameTypeAs, severity(-1))
	})

	Convey("severity values match logging spec", t, func() {
		So(FATAL, ShouldEqual, 0)
		So(ERROR, ShouldEqual, 1)
		So(WARN, ShouldEqual, 2)
		So(INFO, ShouldEqual, 3)
	})
}

func TestSeverityToLevel(t *testing.T) {
	Convey("given a list of severities", t, func() {
		cases := []struct {
			src severity
			exp slog.Level
		}{
			{INFO, slog.LevelInfo},
			{FATAL, LevelFatal},
			{ERROR, slog.LevelError},
			{WARN, slog.LevelWarn},
			{severity(99), slog.LevelInfo},
		}

		for _, tc := range cases {
			Convey(fmt.Sprintf("SeverityToLevel(%v) should equal expected value", tc.src), func() {
				So(SeverityToLevel(tc.src), ShouldEqual, tc.exp)
			})
		}
	})
}

func TestLevelToSeverity(t *testing.T) {
	Convey("given a list of log levels", t, func() {
		cases := []struct {
			src slog.Level
			exp severity
		}{
			{slog.LevelInfo, INFO},
			{LevelFatal, FATAL},
			{slog.LevelError, ERROR},
			{slog.LevelWarn, WARN},
			{slog.LevelDebug, INFO},
			{slog.Level(99), INFO},
		}

		for _, tc := range cases {
			Convey(fmt.Sprintf("LevelToSeverity(%v) should equal expected value", tc.src), func() {
				So(LevelToSeverity(tc.src), ShouldEqual, tc.exp)
			})
		}
	})
}
