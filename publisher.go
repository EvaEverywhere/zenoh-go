package zenoh

// Publisher allows publishing data to a key expression.
//
// Publishers are created via Session.Publisher() and can be reused
// for multiple Put operations. They are automatically closed when
// the session is closed.
//
// Example:
//
//	pub, err := session.Publisher("reachy_mini/command")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Publish multiple times
//	pub.Put([]byte(`{"action": "look_left"}`))
//	pub.Put([]byte(`{"action": "look_right"}`))
type Publisher interface {
	// Put publishes data to the key expression.
	// The data is sent asynchronously.
	Put(data []byte) error

	// Delete publishes a deletion to the key expression.
	// This notifies subscribers that the resource was deleted.
	Delete() error

	// Close releases the publisher resources.
	// After Close, Put and Delete will return errors.
	Close() error
}

