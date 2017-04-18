package influxdb

import "net"

// MaxUDPPayloadSize is a reasonable maximum payload size for UDP packets that
// could be traveling over the internet.
const MaxUDPPayloadSize = 512

var _ Writer = &UDPWriter{}

// UDPWriter is a simple writer that will write points over udp.
type UDPWriter struct {
	conn net.Conn
}

// NewUDPWriter creates a new UDPWriter that will be sent to the specified
// address and will be encoded with the given protocol.
func NewUDPWriter(addr string) (*UDPWriter, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}
	return &UDPWriter{conn: conn}, nil
}

// Write will write the data directly to the UDP socket.
func (w *UDPWriter) Write(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}
	return w.conn.Write(data)
}

// Protocol returns the protocol associated with this UDP writer.
// This will always be the DefaultWriteProtocol.
func (w *UDPWriter) Protocol() Protocol {
	return DefaultWriteProtocol
}

// Close closes the UDP socket.
func (w *UDPWriter) Close() error {
	return w.conn.Close()
}
