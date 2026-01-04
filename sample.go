package zenoh

import "time"

// KeyExpr represents a Zenoh key expression.
//
// Key expressions are strings that identify resources in Zenoh.
// They support wildcards:
//   - "*" matches any single chunk (between slashes)
//   - "**" matches any sequence of chunks
//
// Examples:
//   - "reachy_mini/command" - exact match
//   - "reachy_mini/*" - matches reachy_mini/command, reachy_mini/status, etc.
//   - "reachy_mini/**" - matches all under reachy_mini/
type KeyExpr string

// String returns the key expression as a string.
func (k KeyExpr) String() string {
	return string(k)
}

// Sample represents a received Zenoh sample.
type Sample struct {
	// KeyExpr is the key expression this sample was published to.
	KeyExpr KeyExpr

	// Payload is the raw bytes of the sample.
	Payload []byte

	// Timestamp when the sample was received locally.
	Timestamp time.Time

	// Kind indicates PUT or DELETE.
	Kind SampleKind
}

// String returns the payload as a string.
func (s Sample) String() string {
	return string(s.Payload)
}

// SampleKind indicates the type of sample.
type SampleKind int

const (
	// SampleKindPut indicates a PUT operation (data published)
	SampleKindPut SampleKind = 0

	// SampleKindDelete indicates a DELETE operation
	SampleKindDelete SampleKind = 1
)

// String returns a string representation of the sample kind.
func (k SampleKind) String() string {
	switch k {
	case SampleKindPut:
		return "PUT"
	case SampleKindDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// Handler is called when a sample is received.
// Handlers are invoked in a separate goroutine and should not block for long.
type Handler func(Sample)





