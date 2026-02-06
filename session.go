package zenoh

import (
	"context"
	"fmt"
)

// Session represents a Zenoh session.
// Use Open() to create a session.
type Session interface {
	// Publisher declares a publisher for the given key expression.
	// Publishers can be reused for multiple Put operations.
	Publisher(keyExpr KeyExpr) (Publisher, error)

	// Subscribe creates a subscriber for the given key expression.
	// The handler is called for each received sample.
	// Supports wildcards: "topic/*" or "topic/**"
	Subscribe(keyExpr KeyExpr, handler Handler) (Subscriber, error)

	// Get performs a query and returns matching samples.
	// This is a blocking call that waits for replies.
	Get(ctx context.Context, keyExpr KeyExpr) ([]Sample, error)

	// Close closes the session and all associated resources.
	// After Close, all operations on the session will return ErrSessionClosed.
	Close() error

	// Info returns session information for debugging.
	Info() SessionInfo
}

// SessionInfo contains session metadata.
type SessionInfo struct {
	// ID is the unique session identifier (if available).
	ID string

	// Mode is "peer" or "client".
	Mode string

	// Endpoints this session is connected to.
	Endpoints []string

	// UsingCGO indicates if native bindings are used.
	// False means mock implementation is active.
	UsingCGO bool
}

// Open creates a new Zenoh session with the given configuration.
// Returns an error if the connection fails or configuration is invalid.
//
// Example:
//
//	session, err := zenoh.Open(zenoh.ClientConfig("tcp/192.168.68.80:7447"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close()
func Open(cfg Config) (Session, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return openSession(cfg)
}

// openSession is implemented in session_cgo.go or session_mock.go
// based on build tags.









