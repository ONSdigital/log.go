package log

import (
	"context"
	"log/slog"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestModifyingHandler_Handle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const (
		testEvent = "some event"
		testPC    = 12345
	)
	var (
		testTime = time.Date(2024, time.July, 22, 15, 26, 13, 0, time.UTC)
	)

	Convey("Given a mocked default logger wrapped with a ModifyingHandler", t, func() {
		mockHndlr := &mockHandler{}
		modHandler := ModifyingHandler{mockHndlr}

		Convey("When ", func() {
			mockHndlr.Reset()
			srcRecord := slog.NewRecord(testTime, slog.LevelWarn, testEvent, testPC)
			srcRecord.AddAttrs(
				slog.Int("intKey0", 10),
				slog.String("stringKey0", "string value 0"),
				slog.Group("some_group", slog.Int("intKey1", 11), slog.String("stringKey1", "string value 1")),
				slog.Group("another_group", slog.Int("intKey2", 12), slog.String("stringKey2", "string value 2"),
					slog.Group("sub_group", slog.String("stringKey21", "string value 21"))),
			)
			modHandler.Handle(ctx, srcRecord)

			Convey("The underlying handler should be called", func() {
				So(mockHndlr.handeRecords, ShouldHaveLength, 1)
			})

			record := mockHndlr.handeRecords[0]

			Convey("The record should be passed through to the underlying handler", func() {
				So(record.Time, ShouldEqual, testTime)
				So(record.Level, ShouldEqual, LevelWarn)
				So(record.Message, ShouldResemble, testEvent)
				So(record.PC, ShouldEqual, uintptr(testPC))
			})

			values := getValuesFromRecord(&record)

			Convey("The record should contain values", func() {
				So(values, ShouldNotBeEmpty)
			})

			Convey("The severity should be calculated", func() {
				So(values, ShouldContainKey, "severity")
				value := values["severity"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, WARN)
			})

			Convey("Attrs should be wrapped in a 'data' group", func() {
				So(values, ShouldNotContainKey, "intKey0")
				So(values, ShouldContainKey, "data.intKey0")
				value := values["data.intKey0"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, 10)

				So(values, ShouldNotContainKey, "stringKey0")
				So(values, ShouldContainKey, "data.stringKey0")
				value = values["data.stringKey0"]
				So(value.Kind(), ShouldEqual, slog.KindString)
				So(value.String(), ShouldEqual, "string value 0")

				So(values, ShouldNotContainKey, "some_group.intKey1")
				So(values, ShouldContainKey, "data.some_group.intKey1")
				value = values["data.some_group.intKey1"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, 11)

				So(values, ShouldNotContainKey, "some_group.stringKey1")
				So(values, ShouldContainKey, "data.some_group.stringKey1")
				value = values["data.some_group.stringKey1"]
				So(value.Kind(), ShouldEqual, slog.KindString)
				So(value.String(), ShouldEqual, "string value 1")

				So(values, ShouldNotContainKey, "another_group.intKey2")
				So(values, ShouldContainKey, "data.another_group.intKey2")
				value = values["data.another_group.intKey2"]
				So(value.Kind(), ShouldEqual, slog.KindInt64)
				So(value.Int64(), ShouldEqual, 12)

				So(values, ShouldNotContainKey, "another_group.stringKey2")
				So(values, ShouldContainKey, "data.another_group.stringKey2")
				value = values["data.another_group.stringKey2"]
				So(value.Kind(), ShouldEqual, slog.KindString)
				So(value.String(), ShouldEqual, "string value 2")

				So(values, ShouldNotContainKey, "another_group.sub_group.stringKey21")
				So(values, ShouldContainKey, "data.another_group.sub_group.stringKey21")
				value = values["data.another_group.sub_group.stringKey21"]
				So(value.Kind(), ShouldEqual, slog.KindString)
				So(value.String(), ShouldEqual, "string value 21")
			})
		})
	})
}
