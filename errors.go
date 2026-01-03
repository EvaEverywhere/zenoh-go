package zenoh

import "errors"

// Common errors returned by zenoh-go operations.
var (
	// ErrSessionClosed is returned when operating on a closed session.
	ErrSessionClosed = errors.New("zenoh: session closed")

	// ErrInvalidKeyExpr is returned for malformed key expressions.
	ErrInvalidKeyExpr = errors.New("zenoh: invalid key expression")

	// ErrConnectionFailed is returned when connection to router fails.
	ErrConnectionFailed = errors.New("zenoh: connection failed")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("zenoh: timeout")

	// ErrPublishFailed is returned when a publish operation fails.
	ErrPublishFailed = errors.New("zenoh: publish failed")

	// ErrSubscribeFailed is returned when creating a subscription fails.
	ErrSubscribeFailed = errors.New("zenoh: subscribe failed")

	// ErrQueryFailed is returned when a query operation fails.
	ErrQueryFailed = errors.New("zenoh: query failed")
)

