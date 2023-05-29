package log

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/common"
)

// Each IntermediaryEvent... need their own sync.Pool
var eventBufPool4 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

var eventBufPool5 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

var eventBufPool6 = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

// CustomLogEvent1 for use in middleware to replace the log.Event
func CustomLogEvent1(ctx context.Context, event string,
	req *http.Request, statusCode int, responseContentLength int64,
	startedAt, endedAt *time.Time) {

	buf := eventBufPool4.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	unrollCreatedAt(buf, time.Now().UTC())

	if Namespace != "" {
		buf.WriteByte(',')
		unrollNamespace(buf, Namespace)
	}

	if event != "" {
		buf.WriteByte(',')
		unrollEvent(buf, event)
	}

	if ctx != nil {
		buf.WriteByte(',')
		unrollTraceID(buf, common.GetRequestId(ctx))
	}

	if req != nil {

		// We know what the '*req' is, so its contents can be directly
		// extracted ...
		buf.WriteString(",\"http\":{\"status_code\":")
		unrollInt(buf, statusCode)

		if req.Method != "" {
			buf.WriteString(",\"method\":\"")
			buf.WriteString(req.Method)
			buf.WriteByte('"')
		}

		if req.URL.Path != "" {
			buf.WriteString(",\"path\":\"")
			buf.WriteString(req.URL.Path)
			buf.WriteByte('"')
		}

		if req.URL.RawQuery != "" {
			buf.WriteString(",\"query\":\"")
			buf.WriteString(req.URL.RawQuery)
			buf.WriteByte('"')
		}

		if startedAt != nil {
			buf.WriteString(",\"started_at\":\"")
			unrollTimeToBuf(buf, *startedAt)
			buf.WriteByte('"')
		}

		if endedAt != nil {
			buf.WriteString(",\"ended_at\":\"")
			unrollTimeToBuf(buf, *endedAt)
			buf.WriteByte('"')
		}

		if startedAt != nil && endedAt != nil {
			d := endedAt.Sub(*startedAt)

			buf.WriteString(",\"duration\":")
			unrollInt64(buf, int64(d))
		}

		// We can not easily determine if ResponseContentLength has been assigned a value
		// without doing some sort of reflect which will then create allocs.
		// But if ResponseContentLength is 0, then there is no point showing it.
		if responseContentLength != 0 {
			buf.WriteString(",\"response_content_length\":")
			unrollInt64(buf, responseContentLength)
		}

		buf.WriteByte('}')
	}

	buf.WriteByte('}')
	buf.WriteByte(10)

	l := int64(buf.Len()) // cast to same type as returned by WriteTo()

	// try and write to stdout
	if n, err := buf.WriteTo(destination); n != l || err != nil {
		// if that fails, try and write to stderr
		if n, err := buf.WriteTo(fallbackDestination); n != l || err != nil {
			// if that fails, panic!
			//
			// also defer an os.Exit since the panic might be captured in a recover
			// block in the caller, but we always want to exit in this scenario
			//
			// Note: deferring an os.Exit makes this particular block untestable
			// using conventional `go test`. But it's a narrow enough edge case that
			// it probably isn't worth trying, and only occurs in extreme circumstances
			// (os.Stdout and os.Stderr both being closed) where unpredictable
			// behaviour is expected. It's not clear what a panic or os.Exit would do
			// in this scenario, or if our process is even still alive to get this far.
			defer os.Exit(1)
			panic("error writing log data: " + err.Error())
		}
	}

	eventBufPool4.Put(buf)
}

func unrollProproxyURL(buf *bytes.Buffer, proproxyURL *url.URL) {
	// This function does the equivalent of: json.NewEncoder(buf).Encode(*proproxyURL)
	// but without putting any allocations on the HEAP.

	//////////////////////////////////////////////////////////////////////////////
	// The following comments were copied from go/src/net/url/url.go for reference
	//

	// A URL represents a parsed URL (technically, a URI reference).
	//
	// The general form represented is:
	//
	//	[scheme:][//[userinfo@]host][/]path[?query][#fragment]
	//
	// URLs that do not start with a slash after the scheme are interpreted as:
	//
	//	scheme:opaque[?query][#fragment]
	//
	// Note that the Path field is stored in decoded form: /%47%6f%2f becomes /Go/.
	// A consequence is that it is impossible to tell which slashes in the Path were
	// slashes in the raw URL and which were %2f. This distinction is rarely important,
	// but when it is, the code should use RawPath, an optional field which only gets
	// set if the default encoding is different from Path.
	//
	// URL's String method uses the EscapedPath method to obtain the path. See the
	// EscapedPath method for more details.
	/*type URL struct {
		Scheme     string
		Opaque     string    // encoded opaque data
		User       *Userinfo // username and password information
		Host       string    // host or host:port
		Path       string    // path (relative paths may omit leading slash)
		RawPath    string    // encoded path hint (see EscapedPath method)
		ForceQuery bool      // append a query ('?') even if RawQuery is empty
		RawQuery   string    // encoded query values, without '?'
		Fragment   string    // fragment for references, without '#'
	}*/
	// The Userinfo type is an immutable encapsulation of username and
	// password details for a URL. An existing Userinfo value is guaranteed
	// to have a username set (potentially empty, as allowed by RFC 2396),
	// and optionally a password.
	/*type Userinfo struct {
		username    string
		password    string
		passwordSet bool
	}*/

	buf.WriteString("{\"Scheme\":\"")
	if proproxyURL.Scheme != "" {
		buf.WriteString(proproxyURL.Scheme)
	}
	buf.WriteString("\",\"Opaque\":\"")
	if proproxyURL.Opaque != "" {
		buf.WriteString(proproxyURL.Opaque)
	}
	buf.WriteString("\",\"User\":")
	if proproxyURL.User == nil {
		buf.WriteString("null")
	} else {
		// this section's output format is based on the structure variable names,
		// as i've not seen any examples of how these fields are printed.
		buf.WriteString("{\"Name\":\"")
		buf.WriteString(proproxyURL.User.Username())
		buf.WriteString("\",\"Pass\":\"")
		pass, set := proproxyURL.User.Password()
		buf.WriteString(pass)
		buf.WriteString("\",\"Set\":")
		if set {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		buf.WriteByte('}')
	}
	buf.WriteString(",\"Host\":\"")
	if proproxyURL.Host != "" {
		buf.WriteString(proproxyURL.Host)
	}
	buf.WriteString("\",\"Path\":\"")
	if proproxyURL.Path != "" {
		buf.WriteString(proproxyURL.Path)
	}
	buf.WriteString("\",\"RawPath\":\"")
	if proproxyURL.RawPath != "" {
		buf.WriteString(proproxyURL.RawPath)
	}
	buf.WriteString("\",\"ForceQuery\":")
	if proproxyURL.ForceQuery {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
	buf.WriteString(",\"RawQuery\":\"")
	if proproxyURL.RawQuery != "" {
		buf.WriteString(proproxyURL.RawQuery)
	}
	buf.WriteString("\",\"Fragment\":\"")
	if proproxyURL.Fragment != "" {
		buf.WriteString(proproxyURL.Fragment)
	}
	buf.WriteString("\"}")
}

// CustomLogEvent2 for use in dp-frontend-router createReverseProxy()
// to replace the log.Event
func CustomLogEvent2(ctx context.Context, event string, info severity,
	req *http.Request, statusCode int, responseContentLength int64,
	urlName string, proproxyURL *url.URL,
	proxName string, proxyName string) {

	buf := eventBufPool5.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	unrollCreatedAt(buf, time.Now().UTC())

	if event != "" {
		buf.WriteByte(',')
		unrollEvent(buf, event)
	}

	if ctx != nil {
		buf.WriteByte(',')
		unrollTraceID(buf, common.GetRequestId(ctx))
	}

	buf.WriteByte(',')
	unrollSeverity(buf, int(info))

	if req != nil {

		// We know what the '*req' is, so its contents can be directly
		// extracted ...
		buf.WriteString(",\"http\":{\"status_code\":")
		unrollInt(buf, statusCode)

		if req.URL.RawQuery != "" {
			buf.WriteString(",\"query\":\"")
			buf.WriteString(req.URL.RawQuery)
			buf.WriteByte('"')
		}

		// We can not easily determine if ResponseContentLength has been assigned a value
		// without doing some sort of reflect which will then create allocs.
		// But if ResponseContentLength is 0, then there is no point showing it.
		if responseContentLength != 0 {
			buf.WriteString(",\"response_content_length\":")
			unrollInt64(buf, responseContentLength)
		}

		buf.WriteByte('}')
	}

	if urlName != "" || proxName != "" {
		buf.WriteString(",\"data\":{")
		var somethingWritten bool

		if urlName != "" {
			somethingWritten = true
			buf.WriteByte('"')
			buf.WriteString(urlName)
			buf.WriteString("\":")
			// inline expand the Encode() to save final 128 bytes
			// (as per the commented out structures in unrollProproxyURL() )
			unrollProproxyURL(buf, proproxyURL)
			// json.NewEncoder(buf).Encode(*proproxyURL)
			// buf.Truncate(buf.Len() - 1)               // remove the 'new line', as there is more to append
		}

		if proxName != "" {
			if somethingWritten {
				buf.WriteByte(',')
			}
			somethingWritten = true
			buf.WriteByte('"')
			buf.WriteString(proxName)
			buf.WriteString("\":\"")
			buf.WriteString(proxyName)
			buf.WriteByte('"')

		}
		buf.WriteByte('}')
	}

	buf.WriteByte('}')
	buf.WriteByte(10)

	l := int64(buf.Len()) // cast to same type as returned by WriteTo()

	// try and write to stdout
	if n, err := buf.WriteTo(destination); n != l || err != nil {
		// if that fails, try and write to stderr
		if n, err := buf.WriteTo(fallbackDestination); n != l || err != nil {
			// if that fails, panic!
			//
			// also defer an os.Exit since the panic might be captured in a recover
			// block in the caller, but we always want to exit in this scenario
			//
			// Note: deferring an os.Exit makes this particular block untestable
			// using conventional `go test`. But it's a narrow enough edge case that
			// it probably isn't worth trying, and only occurs in extreme circumstances
			// (os.Stdout and os.Stderr both being closed) where unpredictable
			// behaviour is expected. It's not clear what a panic or os.Exit would do
			// in this scenario, or if our process is even still alive to get this far.
			defer os.Exit(1)
			panic("error writing log data: " + err.Error())
		}
	}

	eventBufPool5.Put(buf)
}

// CustomLogEvent3 for use in middleware to replace the 2nd log.Event
func CustomLogEvent3(ctx context.Context, event string,
	req *http.Request, statusCode int, responseContentLength int64,
	startedAt, endedAt *time.Time) {

	buf := eventBufPool6.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                                // Must reset before each block of usage

	buf.WriteByte('{')
	unrollCreatedAt(buf, time.Now().UTC())

	if event != "" {
		buf.WriteByte(',')
		unrollEvent(buf, event)
	}

	if ctx != nil {
		buf.WriteByte(',')
		unrollTraceID(buf, common.GetRequestId(ctx))
	}

	if req != nil {

		// We know what the '*req' is, so its contents can be directly
		// extracted ...
		buf.WriteString(",\"http\":{\"status_code\":")
		unrollInt(buf, statusCode)

		if startedAt != nil {
			buf.WriteString(",\"started_at\":\"")
			unrollTimeToBuf(buf, *startedAt)
			buf.WriteByte('"')
		}

		if endedAt != nil {
			buf.WriteString(",\"ended_at\":\"")
			unrollTimeToBuf(buf, *endedAt)
			buf.WriteByte('"')
		}

		if startedAt != nil && endedAt != nil {
			d := endedAt.Sub(*startedAt)

			buf.WriteString(",\"duration\":")
			unrollInt64(buf, int64(d))
		}

		// We can not easily determine if ResponseContentLength has been assigned a value
		// without doing some sort of reflect which will then create allocs.
		// But if ResponseContentLength is 0, then there is no point showing it.
		if responseContentLength != 0 {
			buf.WriteString(",\"response_content_length\":")
			unrollInt64(buf, responseContentLength)
		}

		buf.WriteByte('}')
	}

	buf.WriteByte('}')
	buf.WriteByte(10)

	l := int64(buf.Len()) // cast to same type as returned by WriteTo()

	// try and write to stdout
	if n, err := buf.WriteTo(destination); n != l || err != nil {
		// if that fails, try and write to stderr
		if n, err := buf.WriteTo(fallbackDestination); n != l || err != nil {
			// if that fails, panic!
			//
			// also defer an os.Exit since the panic might be captured in a recover
			// block in the caller, but we always want to exit in this scenario
			//
			// Note: deferring an os.Exit makes this particular block untestable
			// using conventional `go test`. But it's a narrow enough edge case that
			// it probably isn't worth trying, and only occurs in extreme circumstances
			// (os.Stdout and os.Stderr both being closed) where unpredictable
			// behaviour is expected. It's not clear what a panic or os.Exit would do
			// in this scenario, or if our process is even still alive to get this far.
			defer os.Exit(1)
			panic("error writing log data: " + err.Error())
		}
	}

	eventBufPool6.Put(buf)
}
