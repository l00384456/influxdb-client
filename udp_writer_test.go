package influxdb_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client"
)

func TestUDPWriter_WritePoint(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	type resp struct {
		out string
		err error
	}
	in := make(chan resp, 5)
	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)

		var out []byte
		if err == nil {
			out = make([]byte, n)
			copy(out, buf[:n])
		}
		in <- resp{out: string(out), err: err}
		close(in)
	}()

	w, err := influxdb.NewUDPWriter(conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	pt := influxdb.Point{
		Name:   "cpu",
		Fields: map[string]interface{}{"value": 5.0},
		Time:   time.Unix(10, 0),
	}
	if _, err := pt.WriteTo(w); err != nil {
		t.Fatalf("unable to write udp message: %s", err)
	}

	timer := time.NewTimer(100 * time.Millisecond)
	select {
	case <-timer.C:
		t.Fatal("no udp message received")
	case r := <-in:
		timer.Stop()

		if r.err != nil {
			t.Fatalf("error reading from udp socket: %s", r.err)
		} else if got, want := r.out, "cpu value=5 10000000000\n"; got != want {
			t.Fatalf("unexpected udp message: got=%q want=%q", got, want)
		}
	}

	timer = time.NewTimer(100 * time.Millisecond)
	select {
	case <-timer.C:
		t.Fatal("unexpected timeout while waiting for udp reader to close")
	case <-in:
		timer.Stop()
	}
}

func TestUDPWriter_MultiplePayloads(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Retrieve the length of writing one point.
	pt := influxdb.Point{
		Name:   "cpu",
		Tags:   []influxdb.Tag{{Key: "host", Value: "server01"}},
		Fields: map[string]interface{}{"value": 5.0},
		Time:   time.Unix(10, 0),
	}
	var buf bytes.Buffer
	if _, err := influxdb.Encode(&buf, &pt); err != nil {
		t.Fatal(err)
	}
	len := buf.Len()

	points := make([]influxdb.Point, (influxdb.MaxUDPPayloadSize/len)+1)
	for i := range points {
		points[i] = pt
	}

	type resp struct {
		out string
		err error
	}
	in := make(chan resp, 5)
	go func() {
		for i := 0; i < 2; i++ {
			buf := make([]byte, 64*1024)
			n, _, err := conn.ReadFromUDP(buf)

			var out []byte
			if err == nil {
				out = make([]byte, n)
				copy(out, buf[:n])
			}
			in <- resp{out: string(out), err: err}
		}
		close(in)
	}()

	w, err := influxdb.NewUDPWriter(conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if _, err := influxdb.WritePoints(w, points); err != nil {
		t.Fatalf("unable to write udp message: %s", err)
	}

	timer := time.NewTimer(100 * time.Millisecond)
	select {
	case <-timer.C:
		t.Fatal("no udp message received")
	case r := <-in:
		timer.Stop()

		if r.err != nil {
			t.Fatalf("error reading from udp socket: %s", r.err)
		}
	}

	timer = time.NewTimer(100 * time.Millisecond)
	select {
	case <-timer.C:
		t.Fatal("no udp message received")
	case r := <-in:
		timer.Stop()

		if r.err != nil {
			t.Fatalf("error reading from udp socket: %s", r.err)
		} else if got, want := r.out, "cpu,host=server01 value=5 10000000000\n"; got != want {
			t.Fatalf("unexpected udp message: got=%q want=%q", got, want)
		}
	}

	timer = time.NewTimer(100 * time.Millisecond)
	select {
	case <-timer.C:
		t.Fatal("unexpected timeout while waiting for udp reader to close")
	case <-in:
		timer.Stop()
	}
}
