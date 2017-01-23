package influxdb_test

import (
	"bytes"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client"
)

func TestTimedWriter(t *testing.T) {
	// Create the initial buffered writer.
	var buf bytes.Buffer
	bw := influxdb.NewBufferedWriter(&buf)

	pt := influxdb.Point{
		Name:   "cpu",
		Tags:   influxdb.Tags{{Key: "host", Value: "server01"}},
		Fields: map[string]interface{}{"value": 5.0},
	}
	if _, err := pt.WriteTo(bw); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// There should be nothing written to the internal buffer.
	if have, want := buf.Len(), 0; have != want {
		t.Fatalf("unexpected buffer length: have=%#v want=%#v", have, want)
	}

	// Wrap in a timed writer and wait until the interval passes.
	w := influxdb.NewTimedWriter(bw, 10*time.Millisecond)
	defer w.Stop()
	<-time.After(20 * time.Millisecond)

	// Data should have been written.
	if have, want := buf.Len(), 26; have != want {
		t.Fatalf("unexpected buffer length: have=%#v want=%#v", have, want)
	}
}
