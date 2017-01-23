package influxdb

import (
	"bufio"
	"io"
)

const (
	defaultBufSize = 4096
)

var _ Writer = &BufferedWriter{}

// BufferedWriter is a custom bufio.BufferedWriter that will never split the
// contents of a Write call between calls to Write. If a Write would cause the
// buffer to be exceeded, it will first flush the contents and then initiate
// the write. This is to avoid splitting line protocol between two separate
// writes.
type BufferedWriter struct {
	bw *bufio.Writer
	p  Protocol
}

// NewBufferedWriter creates a new BufferedWriter. If the io.Writer passed in
// is a Writer, the BufferedWriter copies the Precision from that.
func NewBufferedWriter(w io.Writer) *BufferedWriter {
	return NewBufferedWriterSize(w, defaultBufSize)
}

// NewBufferedWriterSize creates a new BufferedWriter with a size of size.
func NewBufferedWriterSize(w io.Writer, size int) *BufferedWriter {
	protocol := DefaultWriteProtocol
	if w, ok := w.(Writer); ok {
		protocol = w.Protocol()
	}

	return &BufferedWriter{
		bw: bufio.NewWriterSize(w, size),
		p:  protocol,
	}
}

// Protocol returns the Protocol associated with this BufferedWriter.
func (w *BufferedWriter) Protocol() Protocol {
	return w.p
}

func (w *BufferedWriter) Write(p []byte) (n int, err error) {
	if len(p) > w.bw.Available() {
		// Flush the data in the buffer before writing.
		if w.bw.Buffered() != 0 {
			if err = w.bw.Flush(); err != nil {
				return n, err
			}
		}
	}
	return w.bw.Write(p)
}

// Flush causes any bytes buffered to be written to the underlying Writer.
func (w *BufferedWriter) Flush() error {
	return w.bw.Flush()
}
