package influxdb

import "fmt"

// Precision is the requested precision.
type Precision string

const (
	// PrecisionNanosecond represents nanosecond precision.
	PrecisionNanosecond = Precision("ns")

	// PrecisionMicrosecond represents microsecond precision.
	PrecisionMicrosecond = Precision("u")

	// PrecisionMillisecond represents millisecond precision.
	PrecisionMillisecond = Precision("ms")

	// PrecisionSecond represents second precision.
	PrecisionSecond = Precision("s")

	// PrecisionMinute represents minute precision.
	PrecisionMinute = Precision("m")

	// PrecisionHour represents hour precision.
	PrecisionHour = Precision("h")
)

func (p Precision) String() string {
	return string(p)
}

// WithPrecision augments the protocol with the given precision. If the
// protocol cannot be augmented natively, this wraps it in a protocol that will
// truncate the time for any encoded points to the given precision.
func WithPrecision(protocol Protocol, precision Precision) Protocol {
	// Switch on known protocols to avoid reflection.
	switch protocol := protocol.(type) {
	case *lineProtocolV1:
		newP := &lineProtocolV1{}
		if protocol != nil {
			*newP = *protocol
		}
		newP.Precision = precision
		return newP
	default:
		// TODO(jsternberg): Implement the augmented protocol.
		panic(fmt.Sprintf("unsupported protocol type: %T", protocol))
	}
}

// GetPrecision retrieves the precision from a protocol. If the precision could
// not be discovered, this returns an empty string.
func GetPrecision(protocol Protocol) Precision {
	switch protocol := protocol.(type) {
	case *lineProtocolV1:
		if protocol != nil {
			return protocol.Precision
		}
	}
	return PrecisionNanosecond
}
