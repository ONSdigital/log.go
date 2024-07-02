package pretty

import (
	"bufio"
	"github.com/hokaccha/go-prettyjson"
	"io"
)

// NewPrettyWriter wraps the supplied output writer with a filter that converts lines of JSON to pretty-printed lines
// containing syntax highlighting as formatted by the go-prettyjson library. If there are errors in converting the json
// it is simply output as-is. Note, there is a line length limit in the buffer
// (currently 64k defined by https://pkg.go.dev/bufio#pkg-constants)
func NewPrettyWriter(out io.Writer) io.WriteCloser {
	r, wr := io.Pipe()

	go func() {
		defer wr.Close()
		defer r.Close()

		br := bufio.NewScanner(r)
		for br.Scan() {
			raw := []byte(br.Text())
			output, err := prettyjson.Format(raw)
			if err != nil {
				out.Write(append(raw, '\n'))
				continue
			}
			out.Write(append(output, '\n'))
			continue
		}
	}()

	return wr
}
