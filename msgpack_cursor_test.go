package influxdb_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client"
	"github.com/tinylib/msgp/msgp"
)

// EncodeAsMessagePack takes a slice of interfaces and encodes them in their
// message pack representations. This is meant to encode a JSON looking
// structure directly to message pack instead of writing binary files in the
// test file.
func EncodeAsMessagePack(w io.Writer, blobs ...interface{}) {
	writer := msgp.NewWriter(w)

	var encodeValue func(v interface{})
	encodeValue = func(v interface{}) {
		switch val := v.(type) {
		case float64:
			writer.WriteFloat64(val)
		case int:
			writer.WriteInt(val)
		case string:
			writer.WriteString(val)
		case bool:
			writer.WriteBool(val)
		case map[string]string:
			writer.WriteMapHeader(uint32(len(val)))
			for k, v := range val {
				encodeValue(k)
				encodeValue(v)
			}
		case map[string]interface{}:
			writer.WriteMapHeader(uint32(len(val)))
			for k, v := range val {
				encodeValue(k)
				encodeValue(v)
			}
		case []string:
			writer.WriteArrayHeader(uint32(len(val)))
			for _, v := range val {
				encodeValue(v)
			}
		case []interface{}:
			writer.WriteArrayHeader(uint32(len(val)))
			for _, v := range val {
				encodeValue(v)
			}
		case time.Time:
			writer.WriteTime(val)
		default:
			panic(fmt.Sprintf("invalid value type: %T", val))
		}
	}

	for _, blob := range blobs {
		encodeValue(blob)
	}
	writer.Flush()
}

func TestCursor_ResponseError(t *testing.T) {
	var buf bytes.Buffer
	EncodeAsMessagePack(&buf,
		map[string]string{"error": "no database found"},
	)

	_, err := influxdb.NewCursor(ioutil.NopCloser(&buf), "msgpack")
	if err == nil {
		t.Error("expected error")
	} else if have, want := err.Error(), "no database found"; have != want {
		t.Errorf("unexpected error message %q; want %q", have, want)
	}
}

func TestCursor_ResultError(t *testing.T) {
	var buf bytes.Buffer
	EncodeAsMessagePack(&buf,
		map[string]interface{}{"results": 1},
		map[string]interface{}{"id": 0, "error": "expected err"},
	)

	cur, err := influxdb.NewCursor(ioutil.NopCloser(&buf), "msgpack")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = cur.NextSet()
	if err == nil {
		t.Error("expected error")
	} else if e, ok := err.(influxdb.ErrResult); !ok {
		t.Errorf("got error type %T; want %T", err, e)
	} else if e.Err != "expected err" {
		t.Errorf("unexpected error message %q; want %q", e.Err, "expected err")
	}
}

func TestCursor_SeriesError(t *testing.T) {
	var buf bytes.Buffer
	EncodeAsMessagePack(&buf,
		map[string]interface{}{"results": 1},
		map[string]interface{}{"id": 0},
		[]interface{}{1, false},
		map[string]string{"error": "expected err"},
	)

	cur, err := influxdb.NewCursor(ioutil.NopCloser(&buf), "msgpack")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	result, err := cur.NextSet()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = result.NextSeries()
	if err == nil {
		t.Error("expected error")
	} else if e, ok := err.(influxdb.ErrResult); !ok {
		t.Errorf("got error type %T; want %T", err, e)
	} else if e.Err != "expected err" {
		t.Errorf("unexpected error message %q; want %q", e.Err, "expected err")
	}
}
