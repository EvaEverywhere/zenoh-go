//go:build cgo

package zenoh

/*
#include <zenoh.h>
*/
import "C"

import "runtime/cgo"

// cgoSubscriber wraps a native Zenoh subscriber.
type cgoSubscriber struct {
	session *cgoSession
	keyExpr KeyExpr
	handler Handler
	sub     C.z_owned_subscriber_t
	handle  cgo.Handle
	closed  bool
}

func (s *cgoSubscriber) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	// Drop the subscriber
	C.z_subscriber_drop(C.z_subscriber_move(&s.sub))

	// Delete the cgo handle
	s.handle.Delete()

	return nil
}





