package pretty_test

import (
	"bufio"
	"github.com/ONSdigital/log.go/v3/log/pretty"
	"github.com/acarl005/stripansi"
	"io"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestData(t *testing.T) {
	Convey("With a pretty writer", t, func() {
		pir, piw := io.Pipe()
		pw := pretty.NewPrettyWriter(piw)
		sr := bufio.NewScanner(pir)

		defer pw.Close()

		Convey("A JSON line gets pretty-printed", func() {
			src := `{"a":1,"b":"two"}`
			exp := `{` + "\n" + `  "a": 1,` + "\n" + `  "b": "two"` + "\n" + `}`
			pw.Write([]byte(src + "\n"))
			for _, lnexp := range strings.Split(exp, "\n") {
				sr.Scan()
				ln := sr.Text()
				cleanLn := stripansi.Strip(ln)
				So(cleanLn, ShouldEqual, lnexp)
			}
		})

		Convey("Subsequent JSON lines get pretty-printed", func() {
			src := `{"a":1,"b":"two"}` + "\n" + `{"c":3,"d":"four"}`
			exp := `{` + "\n" + `  "a": 1,` + "\n" + `  "b": "two"` + "\n" + `}` + "\n" +
				`{` + "\n" + `  "c": 3,` + "\n" + `  "d": "four"` + "\n" + `}`
			pw.Write([]byte(src + "\n"))
			for _, lnexp := range strings.Split(exp, "\n") {
				sr.Scan()
				ln := sr.Text()
				cleanLn := stripansi.Strip(ln)
				So(cleanLn, ShouldEqual, lnexp)
			}
		})

		Convey("Non-JSON lines get printed as-is", func() {
			src := `Something here`
			exp := `Something here`
			pw.Write([]byte(src + "\n"))
			for _, lnexp := range strings.Split(exp, "\n") {
				sr.Scan()
				ln := sr.Text()
				cleanLn := stripansi.Strip(ln)
				So(cleanLn, ShouldEqual, lnexp)
			}
		})

	})
}
