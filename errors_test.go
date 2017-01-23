package influxdb_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	influxdb "github.com/influxdata/influxdb-client"
)

func TestReadError_JSONError(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: ioutil.NopCloser(strings.NewReader(`{"error":"expected err"}`)),
	}

	if err := influxdb.ReadError(resp); err == nil {
		t.Error("expected error")
	} else if have, want := err.Error(), "expected err"; have != want {
		t.Errorf("unexpected error: have=%#v want=%#v", have, want)
	}
}

func TestReadError_UnknownError(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:       ioutil.NopCloser(strings.NewReader(``)),
		Status:     fmt.Sprintf("%d %s", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)),
		StatusCode: http.StatusInternalServerError,
	}

	if err := influxdb.ReadError(resp); err == nil {
		t.Error("expected error")
	} else if have, want := err.Error(), "unknown http error: 500 Internal Server Error"; have != want {
		t.Errorf("unexpected error: have=%#v want=%#v", have, want)
	}
}
