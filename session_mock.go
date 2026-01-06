//go:build !cgo

package zenoh

import (
	"context"
	"sync"
	"time"
)

// mockSession provides an in-memory implementation for testing.
// Used when CGO is disabled.
type mockSession struct {
	config Config

	mu          sync.RWMutex
	closed      bool
	subscribers map[KeyExpr][]Handler
	messages    []Sample
}

// openSession creates a mock session (no CGO).
func openSession(cfg Config) (Session, error) {
	return &mockSession{
		config:      cfg,
		subscribers: make(map[KeyExpr][]Handler),
	}, nil
}

func (s *mockSession) Publisher(keyExpr KeyExpr) (Publisher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrSessionClosed
	}

	return &mockPublisher{session: s, keyExpr: keyExpr}, nil
}

func (s *mockSession) Subscribe(keyExpr KeyExpr, handler Handler) (Subscriber, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrSessionClosed
	}

	s.subscribers[keyExpr] = append(s.subscribers[keyExpr], handler)
	return &mockSubscriber{session: s, keyExpr: keyExpr, handler: handler}, nil
}

func (s *mockSession) Get(ctx context.Context, keyExpr KeyExpr) ([]Sample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, ErrSessionClosed
	}

	var results []Sample
	for _, msg := range s.messages {
		if matchKeyExpr(keyExpr, msg.KeyExpr) {
			results = append(results, msg)
		}
	}
	return results, nil
}

func (s *mockSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.subscribers = nil
	return nil
}

func (s *mockSession) Info() SessionInfo {
	return SessionInfo{
		ID:        "mock-session",
		Mode:      s.config.Mode,
		Endpoints: s.config.Endpoints,
		UsingCGO:  false,
	}
}

// publish is called by mockPublisher to deliver samples.
func (s *mockSession) publish(keyExpr KeyExpr, data []byte, kind SampleKind) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	sample := Sample{
		KeyExpr:   keyExpr,
		Payload:   data,
		Timestamp: time.Now(),
		Kind:      kind,
	}

	// Store for Get queries
	s.messages = append(s.messages, sample)

	// Notify matching subscribers
	for pattern, handlers := range s.subscribers {
		if matchKeyExpr(pattern, keyExpr) {
			for _, h := range handlers {
				// Call handler in goroutine to avoid blocking
				go h(sample)
			}
		}
	}
}

// matchKeyExpr and matchParts are defined in keyexpr.go








