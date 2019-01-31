package log

import (
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGoLogger(t *testing.T) {
	mock := &eventFuncMock{}
	oldEvent := eventFuncInst
	defer func() {
		eventFuncInst = oldEvent
	}()
	eventFuncInst = &eventFunc{mock.Event}

	Convey("Log data from standard library logger is captured", t, func() {
		So(mock.hasBeenCalled, ShouldBeFalse)
		So(mock.capCtx, ShouldBeNil)
		So(mock.capEvent, ShouldBeEmpty)
		So(mock.capOpts, ShouldHaveLength, 0)

		log.Println("test")

		So(mock.hasBeenCalled, ShouldBeTrue)
		So(mock.capCtx, ShouldBeNil)
		So(mock.capEvent, ShouldEqual, "third party logs")
		So(mock.capOpts, ShouldHaveLength, 1)
		So(mock.capOpts[0], ShouldHaveSameTypeAs, Data{})
		capData := mock.capOpts[0].(Data)
		So(capData, ShouldContainKey, "raw")
		So(capData["raw"], ShouldHaveSameTypeAs, "example")
		So(capData["raw"], ShouldEqual, "test")
	})
}
