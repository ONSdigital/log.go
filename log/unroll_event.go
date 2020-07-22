package log

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"
)

var eventBufPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{} // this is the same as return new(bytes.Buffer)
	},
}

// unrollIntToBuf2 writes lowest 2 characters of int to buffer
func unrollIntToBuf2(buf *bytes.Buffer, value int) {
	c1 := byte((value % 10) + '0')
	c2 := byte(((value / 10) % 10) + '0')
	buf.WriteByte(c2)
	buf.WriteByte(c1)
}

// unrollIntToBuf4 writes lowest 4 characters of int to buffer
func unrollIntToBuf4(buf *bytes.Buffer, value int) {
	c1 := byte((value % 10) + '0')
	value = value / 10
	c2 := byte((value % 10) + '0')
	value = value / 10
	c3 := byte((value % 10) + '0')
	value = value / 10
	c4 := byte((value % 10) + '0')
	buf.WriteByte(c4)
	buf.WriteByte(c3)
	buf.WriteByte(c2)
	buf.WriteByte(c1)
}

// unrollIntToBuf9 converts int to upto 9 digit string with trailling zero's dropped
// but leaving one zero if the number is zero, and tags a 'Z' on the end.
// Assumes the number is positive.
func unrollIntToBuf9(buf *bytes.Buffer, value int) {
	var out [11]byte

	out[8] = byte((value % 10) + '0')
	value = value / 10
	out[7] = byte((value % 10) + '0')
	value = value / 10
	out[6] = byte((value % 10) + '0')
	value = value / 10
	out[5] = byte((value % 10) + '0')
	value = value / 10
	out[4] = byte((value % 10) + '0')
	value = value / 10
	out[3] = byte((value % 10) + '0')
	value = value / 10
	out[2] = byte((value % 10) + '0')
	value = value / 10
	out[1] = byte((value % 10) + '0')
	value = value / 10
	out[0] = byte((value % 10) + '0')

	var last int = 0
	for i := 8; i >= 0; i-- {
		if out[i] != '0' {
			last = i
			break
		}
	}
	// now output all digits, except for any trailing zero's
	for i := 0; i <= last; i++ {
		buf.WriteByte(out[i])
	}
	buf.WriteByte('Z')
}

func unrollTimeToBuf(buf *bytes.Buffer, value time.Time) {
	unrollIntToBuf4(buf, value.Year())
	buf.WriteByte('-')
	unrollIntToBuf2(buf, int(value.Month()))
	buf.WriteByte('-')
	unrollIntToBuf2(buf, int(value.Day()))
	buf.WriteByte('T')
	unrollIntToBuf2(buf, int(value.Hour()))
	buf.WriteByte(':')
	unrollIntToBuf2(buf, int(value.Minute()))
	buf.WriteByte(':')
	unrollIntToBuf2(buf, int(value.Second()))
	buf.WriteByte('.')
	unrollIntToBuf9(buf, int(value.Nanosecond()))
}

// unrollInt writes character version of int to buffer
func unrollInt(buf *bytes.Buffer, n int) {
	var out [15]byte
	var c int

	if n < 0 {
		n = -n
		buf.WriteByte('-')
	}
	for {
		out[c] = byte((n % 10) + '0')
		c++
		n = n / 10
		if n == 0 {
			break
		}
	}

	c--
	for {
		buf.WriteByte(out[c])
		if c == 0 {
			break
		}
		c--
	}
}

// unrollInt64 writes character version of int64 to buffer
func unrollInt64(buf *bytes.Buffer, n int64) {
	var out [25]byte
	var c int

	if n < 0 {
		n = -n
		buf.WriteByte('-')
	}
	for {
		out[c] = byte((n % 10) + '0')
		c++
		n = n / 10
		if n == 0 {
			break
		}
	}

	c--
	for {
		buf.WriteByte(out[c])
		if c == 0 {
			break
		}
		c--
	}
}

func unrollCreatedAt(buf *bytes.Buffer, value time.Time) {
	buf.WriteByte('"')
	buf.WriteString("created_at")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte('"')
	unrollTimeToBuf(buf, value)
	buf.WriteByte('"')
}

func unrollNamespace(buf *bytes.Buffer, value string) {
	buf.WriteByte('"')
	buf.WriteString("namespace")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte('"')
	buf.WriteString(value)
	buf.WriteByte('"')
}

func unrollEvent(buf *bytes.Buffer, value string) {
	buf.WriteByte('"')
	buf.WriteString("event")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte('"')
	buf.WriteString(value)
	buf.WriteByte('"')
}

func unrollTraceId(buf *bytes.Buffer, value string) {
	buf.WriteByte('"')
	buf.WriteString("trace_id")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte('"')
	buf.WriteString(value)
	buf.WriteByte('"')
}

func unrollSeverity(buf *bytes.Buffer, value int) {
	buf.WriteByte('"')
	buf.WriteString("severity")
	buf.WriteByte('"')
	buf.WriteByte(':')
	unrollInt(buf, value)
}

func unrollHTTPToBuf(buf *bytes.Buffer, value *EventHTTP) {
	// We know what the '*EventHTTP' is, so its contents can be directly
	// extracted ...
	buf.WriteByte('"')
	buf.WriteString("http")
	buf.WriteByte('"')
	buf.WriteByte(':')

	buf.WriteByte('{')

	var somethingWritten bool
	if value.StatusCode != nil {
		buf.WriteByte('"')
		buf.WriteString("status_code")
		buf.WriteByte('"')
		buf.WriteByte(':')
		unrollInt(buf, *value.StatusCode)
		somethingWritten = true
	}

	if value.Method != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("method")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Method)
		buf.WriteByte('"')
	}

	if value.Scheme != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("scheme")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Scheme)
		buf.WriteByte('"')
	}

	if value.Host != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("host")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Host)
		buf.WriteByte('"')
	}

	// port will always have some value ...
	if somethingWritten {
		buf.WriteByte(',')
	}
	somethingWritten = true
	buf.WriteByte('"')
	buf.WriteString("port")
	buf.WriteByte('"')
	buf.WriteByte(':')
	unrollInt(buf, value.Port)

	if value.Path != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("path")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Path)
		buf.WriteByte('"')
	}

	if value.Query != "" {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("query")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		buf.WriteString(value.Query)
		buf.WriteByte('"')
	}

	if value.StartedAt != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("started_at")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		unrollTimeToBuf(buf, *value.StartedAt)
		buf.WriteByte('"')
	}

	if value.EndedAt != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("ended_at")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('"')
		unrollTimeToBuf(buf, *value.EndedAt)
		buf.WriteByte('"')
	}

	if value.Duration != nil {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true
		buf.WriteByte('"')
		buf.WriteString("duration")
		buf.WriteByte('"')
		buf.WriteByte(':')
		unrollInt64(buf, int64(*value.Duration))
	}

	// We can not easily determine if ResponseContentLength has been assigned a value
	// without doing some sort of reflect which will then create allocs.
	// But if ResponseContentLength is 0, then there is no point showing it.
	if value.ResponseContentLength != 0 {
		if somethingWritten {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString("response_content_length")
		buf.WriteByte('"')
		buf.WriteByte(':')
		unrollInt64(buf, int64(value.ResponseContentLength))
	}

	buf.WriteByte('}')
}

func unrollAuthToBuf(somethingWritten bool, buf *bytes.Buffer, value *eventAuth) {
	if somethingWritten {
		buf.WriteByte(',')
	}
	if value.Identity != "" || value.IdentityType != "" {
		var somethingNew bool

		buf.WriteByte('"')
		buf.WriteString("auth")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte('{')

		if value.Identity != "" {

			buf.WriteByte('"')
			buf.WriteString("identity")
			buf.WriteByte('"')
			buf.WriteByte(':')
			buf.WriteByte('"')
			buf.WriteString(value.Identity)
			buf.WriteByte('"')

			somethingNew = true
		}
		if value.IdentityType != "" {
			if somethingNew {
				buf.WriteByte(',')
			}

			buf.WriteByte('"')
			buf.WriteString("identity_type")
			buf.WriteByte('"')
			buf.WriteByte(':')
			buf.WriteByte('"')
			buf.WriteString(string(value.IdentityType))
			buf.WriteByte('"')
		}
		buf.WriteByte('}')
	} else {
		buf.WriteByte('{')
		buf.WriteByte('}')
	}
}

func unrollDataToBuf(buf *bytes.Buffer, value *Data) {
	// We know what the '*Data' is 'map[string]interface{}'
	// and we know that the three Event() calls on the HOT-PATH in dp-frontend-router
	// only pass a string or URL for the value ... so with this prior knowledge ...
	// to achieve the goal of minimum memory allocations, this function will
	// decode our 'knowns' in an allocation optimum way.
	// It will deal with unknowns with an inefficient "json.NewEncoder(buf).Encode(value)"
	// ... that said, it appears that the Encode does not create any allocations on
	// the HEAP for the URL (possibly because it has no sub structure structures or
	// interface{} ?)
	// ODD'ly: sometimes the 'proxy_name' is output before the 'destination' - go figure ?
	var somethingWritten bool

	buf.WriteByte('"')
	buf.WriteString("data")
	buf.WriteByte('"')
	buf.WriteByte(':')

	buf.WriteByte('{')
	for k, v := range *value {
		if somethingWritten {
			buf.WriteByte(',')
		}
		somethingWritten = true

		buf.WriteByte('"')
		buf.WriteString(k) // add the key
		buf.WriteByte('"')
		buf.WriteByte(':')
		switch n := v.(type) {
		case string:
			buf.WriteByte('"')
			buf.WriteString(n) // add the 'known' value type of 'string'
			buf.WriteByte('"')
		default: // too many other possibilities, so use Encode()
			json.NewEncoder(buf).Encode(v)
			buf.Truncate(buf.Len() - 1) // remove the 'new line', as there is more to append
		}
	}
	buf.WriteByte('}')
}

func unrollErrorToBuf(buf *bytes.Buffer, value *EventError) {
	// Error's are not on the HAPPY HOT-PATH of dp-frontend-router, so
	// they don't need unrolling - just use the library function.
	// Also Error's should only be seen outside of production server ...
	buf.WriteByte('"')
	buf.WriteString("error")
	buf.WriteByte('"')
	buf.WriteByte(':')

	json.NewEncoder(buf).Encode(value)
	buf.Truncate(buf.Len() - 1) // remove the 'new line', as there is more to append
}
