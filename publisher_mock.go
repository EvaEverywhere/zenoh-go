//go:build !cgo

package zenoh

// mockPublisher implements Publisher for testing.
type mockPublisher struct {
	session *mockSession
	keyExpr KeyExpr
	closed  bool
}

func (p *mockPublisher) Put(data []byte) error {
	if p.closed {
		return ErrSessionClosed
	}
	p.session.publish(p.keyExpr, data, SampleKindPut)
	return nil
}

func (p *mockPublisher) Delete() error {
	if p.closed {
		return ErrSessionClosed
	}
	p.session.publish(p.keyExpr, nil, SampleKindDelete)
	return nil
}

func (p *mockPublisher) Close() error {
	p.closed = true
	return nil
}

