package influxdb

import "time"

var _ Writer = &TimedWriter{}

// TimedWriter wraps a BufferedWriter and automatically flushes it after each
// duration passes.
type TimedWriter struct {
	*BufferedWriter
	ticker *time.Ticker
	done   chan struct{}
}

// NewTimedWriter creates a new TimedWriter from a BufferedWriter and will
// flush every d time.Duration. After the TimedWriter is created, it must be
// stopped using Stop or this will leak a goroutine that will run infinitely.
func NewTimedWriter(bw *BufferedWriter, d time.Duration) *TimedWriter {
	w := &TimedWriter{
		BufferedWriter: bw,
		ticker:         time.NewTicker(d),
		done:           make(chan struct{}),
	}
	go w.loop()
	return w
}

// Stop stops this TimedWriter. This method will panic if called multiple
// times.
func (w *TimedWriter) Stop() {
	w.ticker.Stop()
	close(w.done)
}

// loop will automatically flush the BufferedWriter each interval or until the
// TimedWriter is stopped.
func (w *TimedWriter) loop() {
	for {
		select {
		case <-w.ticker.C:
			// TODO(jsternberg): Add some way to perform logging.
			w.Flush()
		case <-w.done:
			return
		}
	}
}
