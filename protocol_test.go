package influxdb_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client"
)

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	pt := influxdb.Point{
		Name: "cpu",
		Tags: []influxdb.Tag{
			{Key: "host", Value: "server01"},
		},
		Fields: map[string]interface{}{
			"value": float64(5),
		},
		Time: time.Unix(2, 0),
	}

	if _, err := influxdb.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 2000000000\n"; have != want {
		t.Fatalf("unexpected protocol output: have=%#v want=%#v", have, want)
	}
}

func TestLineProtocol_V1_Encode(t *testing.T) {
	var buf bytes.Buffer
	p := influxdb.LineProtocol.V1()

	pt := influxdb.Point{
		Name: "cpu",
		Fields: map[string]interface{}{
			"value": float64(5),
		},
	}

	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := buf.String(), "cpu value=5\n"; got != want {
		t.Errorf("unexpected protocol output:\n\ngot=%v\nwant=%v\n", got, want)
	}

	buf.Reset()
	pt.Tags = []influxdb.Tag{{Key: "host", Value: "server01"}}

	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := buf.String(), "cpu,host=server01 value=5\n"; got != want {
		t.Errorf("unexpected protocol output:\n\ngot=%v\nwant=%v\n", got, want)
	}

	buf.Reset()
	pt.Time = time.Unix(0, 1000)

	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := buf.String(), "cpu,host=server01 value=5 1000\n"; got != want {
		t.Errorf("unexpected protocol output:\n\ngot=%v\nwant=%v\n", got, want)
	}
}

func TestLineProtocol_V1_ErrNoFields(t *testing.T) {
	var buf bytes.Buffer
	p := influxdb.LineProtocol.V1()

	pt := influxdb.Point{Name: "cpu"}
	if _, err := p.Encode(&buf, &pt); err != influxdb.ErrNoFields {
		t.Fatal("unexpected error: %v", err)
	}
}

func TestLineProtocol_V1_WithPrecision(t *testing.T) {
	var buf bytes.Buffer
	p := influxdb.LineProtocol.V1()

	pt := influxdb.Point{
		Name: "cpu",
		Tags: []influxdb.Tag{
			{Key: "host", Value: "server01"},
		},
		Fields: map[string]interface{}{
			"value": float64(5),
		},
		Time: time.Unix(7265, 769142873),
	}

	p = influxdb.WithPrecision(p, influxdb.PrecisionNanosecond)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 7265769142873\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	p = influxdb.WithPrecision(p, influxdb.PrecisionMicrosecond)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 7265769142\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	p = influxdb.WithPrecision(p, influxdb.PrecisionMillisecond)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 7265769\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	p = influxdb.WithPrecision(p, influxdb.PrecisionSecond)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 7265\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	p = influxdb.WithPrecision(p, influxdb.PrecisionMinute)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 121\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	p = influxdb.WithPrecision(p, influxdb.PrecisionHour)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu,host=server01 value=5 2\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()
}

func TestLineProtocol_V1_Fields(t *testing.T) {
	var buf bytes.Buffer
	p := influxdb.LineProtocol.V1()

	pt := influxdb.Point{
		Name:   "cpu",
		Fields: map[string]interface{}{},
	}

	pt.Fields["value"] = float64(5)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=5\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = float32(5)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=5\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = int64(5)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=5i\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = int32(5)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=5i\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = int(5)
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=5i\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = "foobar"
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=\"foobar\"\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = true
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=t\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = false
	if _, err := p.Encode(&buf, &pt); err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if have, want := buf.String(), "cpu value=f\n"; have != want {
		t.Errorf("unexpected output: have=%#v want=%#v", have, want)
	}
	buf.Reset()

	pt.Fields["value"] = struct{}{}
	if _, err := p.Encode(&buf, &pt); err == nil {
		t.Error("expected error")
	} else if have, want := err.Error(), "invalid field type: struct {}"; have != want {
		t.Errorf("unexpected error: have=%#v want=%#v", have, want)
	}
	buf.Reset()
}

func BenchmarkLineProtocol_V1(b *testing.B) {
	pt := influxdb.Point{
		Name: "cpu",
		Tags: []influxdb.Tag{
			{Key: "host", Value: "server01"},
		},
		Fields: map[string]interface{}{
			"value": float64(5),
		},
		Time: time.Unix(25, 0),
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pt.WriteTo(ioutil.Discard)
	}
}
