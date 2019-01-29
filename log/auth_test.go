package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuth(t *testing.T) {
	Convey("Auth function returns a *eventAuth", t, func() {
		auth := Auth(USER, "test")
		So(auth, ShouldHaveSameTypeAs, &eventAuth{})
		So(auth, ShouldImplement, (*option)(nil))

		Convey("*eventAuth has the correct fields", func() {
			ea := Auth(USER, "test1").(*eventAuth)
			So(ea.IdentityType, ShouldEqual, USER)
			So(ea.Identity, ShouldEqual, "test1")

			ea = Auth(SERVICE, "test2").(*eventAuth)
			So(ea.IdentityType, ShouldEqual, SERVICE)
			So(ea.Identity, ShouldEqual, "test2")
		})
	})

	Convey("*eventAuth can be attached to *EventData", t, func() {
		event := &EventData{}
		So(event.Auth, ShouldBeNil)

		auth := &eventAuth{}
		auth.attach(event)

		So(event.Auth, ShouldEqual, auth)
	})

	Convey("SERVICE stringifies to 'service'", t, func() {
		So(string(SERVICE), ShouldEqual, "service")
	})

	Convey("USER stringifies to 'user'", t, func() {
		So(string(USER), ShouldEqual, "user")
	})
}
