package zenoh

// Subscriber represents an active subscription.
//
// Subscribers are created via Session.Subscribe() and receive samples
// via the provided Handler callback. They are automatically closed
// when the session is closed.
//
// Example:
//
//	sub, err := session.Subscribe("reachy_mini/joint_positions", func(s zenoh.Sample) {
//	    log.Printf("Received: %s", s.Payload)
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sub.Close()
type Subscriber interface {
	// Close stops the subscription and releases resources.
	// After Close, no more samples will be delivered to the handler.
	Close() error
}





