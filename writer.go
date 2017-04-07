package influxdb

import (
	"bytes"
	"io"
	"net/url"
	"strings"
)

// WriteOptions is a set of configuration options for configuring writers.
type WriteOptions struct {
	Database        string
	RetentionPolicy string
	Consistency     Consistency
	Protocol        Protocol
}

// Clone creates a copy of the WriteOptions.
func (opt *WriteOptions) Clone() WriteOptions {
	return *opt
}

// Writer is an interface for something that can write points to somewhere.
// The Writer wraps io.Writer and also exposes the Protocol that points should
// be encoded with when writing to this Writer.
type Writer interface {
	io.Writer
	Protocol() Protocol
}

// WritePoints encodes and writes the points to the passed in Writer using
// either the Protocol returned by the Writer or the default line protocol if
// writing to an io.Writer.
//
// After writing all of the points, the Flush method is called on the io.Writer
// if it supports that method.
func WritePoints(w io.Writer, points []Point) (n int, err error) {
	if len(points) == 0 {
		return 0, nil
	}

	for _, pt := range points {
		c, err := pt.WriteTo(w)
		if err != nil {
			return n, err
		}
		n += c
	}

	type flusher interface {
		Flush() error
	}
	if w, ok := w.(flusher); ok {
		if err := w.Flush(); err != nil {
			return n, err
		}
	}
	return
}

var _ Writer = &HTTPWriter{}

// HTTPWriter holds onto write options and acts as a convenience method for performing writes.
type HTTPWriter struct {
	c *Client
	WriteOptions
}

// Write writes the bytes to the server. The data should be in the line
// protocol format specified in the WriteOptions attached to this writer so the
// server understands the format. Each call to Write will make a single HTTP
// write request.
func (w *HTTPWriter) Write(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	values := url.Values{}
	if w.Database != "" {
		values.Set("db", w.Database)
	}
	if w.RetentionPolicy != "" {
		values.Set("rp", w.RetentionPolicy)
	}
	if consistency := w.Consistency.String(); consistency != "" {
		values.Set("consistency", consistency)
	}
	if precision := GetPrecision(w.Protocol()); precision != "" {
		values.Set("precision", precision.String())
	}

	u := w.c.url("/write")
	u.RawQuery = values.Encode()

	req := newRequest("POST", u.String(), bytes.NewReader(data))
	p := w.Protocol()
	if p == nil {
		p = DefaultWriteProtocol
	}
	req.Header.Set("Content-Type", p.ContentType())
	if w.c.Auth != nil {
		req.SetBasicAuth(w.c.Auth.Username, w.c.Auth.Password)
	}

	resp, err := w.c.Do(req)
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode / 100 {
	case 2:
		return len(data), nil
	case 4:
		// This is a client error. Read the error message to learn what type of
		// error this is.
		err := ReadError(resp)
		if strings.HasPrefix(err.Error(), "partial write:") {
			// So we DID write, but it was a partial write. Wrap the error message.
			return len(data), ErrPartialWrite{Err: err.Error()}
		}
		return 0, err
	default:
		// The server should never actually return anything other than the
		// above, but catch any weird status codes that might get thrown by a
		// proxy or something.
		return 0, ReadError(resp)
	}
}

// Protocol is the Protocol that will be used to write to the HTTP endpoint.
func (w *HTTPWriter) Protocol() Protocol {
	return w.WriteOptions.Protocol
}
