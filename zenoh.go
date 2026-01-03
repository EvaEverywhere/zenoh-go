// Package zenoh provides Go bindings for Eclipse Zenoh.
//
// This package uses CGO to wrap the zenoh-c library. When CGO is disabled
// or zenoh-c is not available, a mock implementation is used for testing.
//
// Basic usage:
//
//	session, err := zenoh.Open(zenoh.ClientConfig("tcp/192.168.68.80:7447"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close()
//
//	// Publish
//	pub, _ := session.Publisher("my/topic")
//	pub.Put([]byte("hello"))
//
//	// Subscribe
//	session.Subscribe("my/topic", func(s zenoh.Sample) {
//	    fmt.Printf("Received: %s\n", s.Payload)
//	})
package zenoh

// Version of the zenoh-go bindings
const Version = "0.1.0"

// Mode constants for session configuration
const (
	// ModePeer operates as a Zenoh peer, can use multicast discovery
	ModePeer = "peer"

	// ModeClient requires explicit endpoints to connect to
	ModeClient = "client"
)

