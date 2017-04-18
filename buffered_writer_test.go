package influxdb_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	influxdb "github.com/influxdata/influxdb-client"
)

func TestBufferedWriter(t *testing.T) {
	calls := 0
	protocol := influxdb.DefaultWriteProtocol
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Method = %q; want %q", got, want)
		}

		values := r.URL.Query()
		if got, want := values.Get("db"), "db0"; got != want {
			t.Errorf("db = %q; want %q", got, want)
		}
		if got, want := values.Get("rp"), "rp0"; got != want {
			t.Errorf("rp = %q; want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), protocol.ContentType(); got != want {
			t.Errorf("Content-Type = %q; want %q", got, want)
		}

		data, _ := ioutil.ReadAll(r.Body)
		if got, want := string(data), "cpu,host=server01 value=5\n"; got != want {
			t.Errorf("body = %q; want %q", got, want)
		}
		calls++

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := influxdb.NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	writer := client.Writer()
	writer.Database = "db0"
	writer.RetentionPolicy = "rp0"
	bw := influxdb.NewBufferedWriterSize(writer, 32)

	pt := influxdb.Point{
		Name:   "cpu",
		Tags:   influxdb.Tags{{Key: "host", Value: "server01"}},
		Fields: map[string]interface{}{"value": 5.0},
	}
	if _, err := pt.WriteTo(bw); err != nil {
		t.Fatal(err)
	} else if have, want := calls, 0; have != want {
		t.Fatalf("invalid number of write calls: have=%#v want=%#v", have, want)
	}
	if _, err := pt.WriteTo(bw); err != nil {
		t.Fatal(err)
	} else if have, want := calls, 1; have != want {
		t.Fatalf("invalid number of write calls: have=%#v want=%#v", have, want)
	}
	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	} else if have, want := calls, 2; have != want {
		t.Fatalf("invalid number of write calls: have=%#v want=%#v", have, want)
	}
}

func TestWritePoints_BufferedWriter(t *testing.T) {
	// Create the initial buffered writer.
	var buf bytes.Buffer
	w := influxdb.NewBufferedWriter(&buf)

	pt := influxdb.Point{
		Name:   "cpu",
		Tags:   influxdb.Tags{{Key: "host", Value: "server01"}},
		Fields: map[string]interface{}{"value": 5.0},
	}
	if _, err := influxdb.WritePoints(w, []influxdb.Point{pt}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// There should be nothing written to the internal buffer.
	if have, want := buf.Len(), 26; have != want {
		t.Fatalf("unexpected buffer length: have=%#v want=%#v", have, want)
	}
}
