//go:build !cgo

package zenoh

// mockSubscriber implements Subscriber for testing.
type mockSubscriber struct {
	session *mockSession
	keyExpr KeyExpr
	handler Handler
	closed  bool
}

func (s *mockSubscriber) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	// Remove handler from session
	s.session.mu.Lock()
	defer s.session.mu.Unlock()

	handlers := s.session.subscribers[s.keyExpr]
	for i, h := range handlers {
		// Compare function pointers (this is a simplification)
		if &h == &s.handler {
			s.session.subscribers[s.keyExpr] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}









