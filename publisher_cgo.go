//go:build cgo

package zenoh

/*
#include <zenoh.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// cgoPublisher wraps a native Zenoh publisher.
type cgoPublisher struct {
	session *cgoSession
	keyExpr KeyExpr
	pub     C.z_owned_publisher_t
	closed  bool
}

func (p *cgoPublisher) Put(data []byte) error {
	if p.closed {
		return ErrSessionClosed
	}

	if len(data) == 0 {
		return p.putEmpty()
	}

	// Create bytes from buffer
	var payload C.z_owned_bytes_t
	C.z_bytes_copy_from_buf(
		&payload,
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	)

	// Put
	result := C.z_publisher_put(
		C.z_publisher_loan(&p.pub),
		C.z_bytes_move(&payload),
		nil,
	)

	if result < 0 {
		return fmt.Errorf("%w: error code %d", ErrPublishFailed, result)
	}

	return nil
}

func (p *cgoPublisher) putEmpty() error {
	var payload C.z_owned_bytes_t
	C.z_bytes_empty(&payload)

	result := C.z_publisher_put(
		C.z_publisher_loan(&p.pub),
		C.z_bytes_move(&payload),
		nil,
	)

	if result < 0 {
		return fmt.Errorf("%w: error code %d", ErrPublishFailed, result)
	}

	return nil
}

func (p *cgoPublisher) Delete() error {
	if p.closed {
		return ErrSessionClosed
	}

	result := C.z_publisher_delete(
		C.z_publisher_loan(&p.pub),
		nil,
	)
	if result < 0 {
		return fmt.Errorf("%w: delete error code %d", ErrPublishFailed, result)
	}
	return nil
}

func (p *cgoPublisher) Close() error {
	// Publisher cleanup is handled by session.Close()
	// Mark as closed to prevent further operations
	p.closed = true
	return nil
}








