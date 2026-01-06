//go:build cgo

package zenoh

/*
#cgo CFLAGS: -I/opt/homebrew/include -I/usr/local/include -I/usr/include
#cgo LDFLAGS: -L/opt/homebrew/lib -L/usr/local/lib -L/usr/lib -lzenohc

#include <zenoh.h>
#include <stdlib.h>
#include <string.h>

// Forward declaration for Go callback (signature must match exactly)
extern void goSampleCallback(struct z_loaned_sample_t*, void*);

// Callback wrapper that C can call
static void sample_callback_wrapper(struct z_loaned_sample_t* sample, void* context) {
    goSampleCallback(sample, context);
}

// Helper to create closure with our wrapper
static z_owned_closure_sample_t make_sample_closure(void* context) {
    z_owned_closure_sample_t closure;
    z_closure_sample(&closure, sample_callback_wrapper, NULL, context);
    return closure;
}

// Helper to create config from JSON5 string
// Returns 0 on success, negative on error
static int config_from_json5(z_owned_config_t* config, const char* json5) {
    // zenoh-c 1.x uses zc_config_from_str
    z_result_t result = zc_config_from_str(config, json5);
    return (int)result;
}

// Helper to insert JSON5 into existing config
// This is an alternative approach if z_config_from_str doesn't work
static int config_insert_json5(z_loaned_config_t* config, const char* key, const char* json5) {
    // zenoh-c 1.0 uses zc_config_insert_json5
    return (int)zc_config_insert_json5(config, key, json5);
}

// Helper to safely get string pointer and length from z_view_string_t
// This abstracts the field names which may vary between zenoh-c versions
static const char* view_string_data(const z_view_string_t* s) {
    return z_string_data(z_view_string_loan(s));
}

static size_t view_string_len(const z_view_string_t* s) {
    return z_string_len(z_view_string_loan(s));
}

// Helper to read bytes from z_bytes into a buffer
// Returns number of bytes read
static size_t read_bytes_to_buffer(const z_loaned_bytes_t* bytes, uint8_t* buffer, size_t buffer_len) {
    // zenoh-c 1.x: z_bytes_get_reader returns reader value
    z_bytes_reader_t reader = z_bytes_get_reader(bytes);
    return z_bytes_reader_read(&reader, buffer, buffer_len);
}
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"runtime/cgo"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// cgoSession wraps a native Zenoh session.
type cgoSession struct {
	session C.z_owned_session_t
	config  Config

	mu          sync.Mutex
	closed      bool
	publishers  map[KeyExpr]*cgoPublisher
	subscribers []*cgoSubscriber
}

// openSession creates a CGO-backed session.
func openSession(cfg Config) (Session, error) {
	var zconfig C.z_owned_config_t

	// Configure based on mode
	if cfg.Mode == ModeClient && len(cfg.Endpoints) > 0 {
		// Build full JSON5 config for client mode
		// Format: { "mode": "client", "connect": { "endpoints": ["tcp/..."] } }
		quotedEndpoints := make([]string, len(cfg.Endpoints))
		for i, ep := range cfg.Endpoints {
			quotedEndpoints[i] = `"` + ep + `"`
		}
		configJSON := fmt.Sprintf(`{"mode":"client","connect":{"endpoints":[%s]}}`,
			strings.Join(quotedEndpoints, ","))

		cJSON := C.CString(configJSON)
		defer C.free(unsafe.Pointer(cJSON))

		// Create config from JSON5
		result := C.config_from_json5(&zconfig, cJSON)
		if result < 0 {
			// Fallback: try default config with manual endpoint insertion
			C.z_config_default(&zconfig)

			// Try inserting connect endpoints
			connectJSON := fmt.Sprintf(`{"endpoints":[%s]}`, strings.Join(quotedEndpoints, ","))
			cConnectJSON := C.CString(connectJSON)
			cConnectKey := C.CString("connect")
			defer C.free(unsafe.Pointer(cConnectJSON))
			defer C.free(unsafe.Pointer(cConnectKey))

			insertResult := C.config_insert_json5(C.z_config_loan(&zconfig), cConnectKey, cConnectJSON)
			if insertResult < 0 {
				// Log warning but continue - may work with defaults
				// The router might be on localhost:7447
			}
		}
	} else {
		// Peer mode or no endpoints - use defaults
		C.z_config_default(&zconfig)
	}

	// Open session
	var session C.z_owned_session_t
	result := C.z_open(&session, C.z_config_move(&zconfig), nil)
	if result < 0 {
		return nil, fmt.Errorf("%w: error code %d", ErrConnectionFailed, result)
	}

	s := &cgoSession{
		session:    session,
		config:     cfg,
		publishers: make(map[KeyExpr]*cgoPublisher),
	}

	// Set finalizer for safety
	runtime.SetFinalizer(s, func(s *cgoSession) {
		s.Close()
	})

	return s, nil
}

func (s *cgoSession) Publisher(keyExpr KeyExpr) (Publisher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrSessionClosed
	}

	// Check cache
	if pub, ok := s.publishers[keyExpr]; ok {
		return pub, nil
	}

	// Create key expression
	cKeyExpr := C.CString(string(keyExpr))
	defer C.free(unsafe.Pointer(cKeyExpr))

	var ke C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&ke, cKeyExpr) < 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidKeyExpr, keyExpr)
	}

	// Declare publisher
	var pub C.z_owned_publisher_t
	result := C.z_declare_publisher(
		C.z_session_loan(&s.session),
		&pub,
		C.z_view_keyexpr_loan(&ke),
		nil,
	)
	if result < 0 {
		return nil, fmt.Errorf("%w for %s: error code %d", ErrPublishFailed, keyExpr, result)
	}

	p := &cgoPublisher{
		session: s,
		keyExpr: keyExpr,
		pub:     pub,
	}

	s.publishers[keyExpr] = p
	return p, nil
}

func (s *cgoSession) Subscribe(keyExpr KeyExpr, handler Handler) (Subscriber, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrSessionClosed
	}

	// Create key expression
	cKeyExpr := C.CString(string(keyExpr))
	defer C.free(unsafe.Pointer(cKeyExpr))

	var ke C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&ke, cKeyExpr) < 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidKeyExpr, keyExpr)
	}

	// Create subscriber wrapper with cgo handle
	sub := &cgoSubscriber{
		session: s,
		keyExpr: keyExpr,
		handler: handler,
	}
	sub.handle = cgo.NewHandle(sub)

	// Create closure with our callback wrapper
	closure := C.make_sample_closure(unsafe.Pointer(sub.handle))

	// Declare subscriber
	result := C.z_declare_subscriber(
		C.z_session_loan(&s.session),
		&sub.sub,
		C.z_view_keyexpr_loan(&ke),
		C.z_closure_sample_move(&closure),
		nil,
	)
	if result < 0 {
		sub.handle.Delete()
		return nil, fmt.Errorf("%w for %s: error code %d", ErrSubscribeFailed, keyExpr, result)
	}

	s.subscribers = append(s.subscribers, sub)
	return sub, nil
}

func (s *cgoSession) Get(ctx context.Context, keyExpr KeyExpr) ([]Sample, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrSessionClosed
	}
	s.mu.Unlock()

	// TODO: Implement z_get with reply handling
	// This requires more complex callback handling for replies
	return nil, fmt.Errorf("%w: Get not yet implemented", ErrQueryFailed)
}

func (s *cgoSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	// Close all subscribers
	for _, sub := range s.subscribers {
		C.z_subscriber_drop(C.z_subscriber_move(&sub.sub))
		sub.handle.Delete()
	}
	s.subscribers = nil

	// Close all publishers
	for _, pub := range s.publishers {
		C.z_publisher_drop(C.z_publisher_move(&pub.pub))
	}
	s.publishers = nil

	// Close session
	C.z_session_drop(C.z_session_move(&s.session))

	// Clear finalizer
	runtime.SetFinalizer(s, nil)

	return nil
}

func (s *cgoSession) Info() SessionInfo {
	return SessionInfo{
		ID:        "cgo-session", // Could extract from z_info_zid
		Mode:      s.config.Mode,
		Endpoints: s.config.Endpoints,
		UsingCGO:  true,
	}
}

//export goSampleCallback
func goSampleCallback(sample *C.z_loaned_sample_t, context unsafe.Pointer) {
	h := cgo.Handle(context)
	sub := h.Value().(*cgoSubscriber)

	// Extract key expression using safe accessor
	keyexpr := C.z_sample_keyexpr(sample)
	var keystr C.z_view_string_t
	C.z_keyexpr_as_view_string(keyexpr, &keystr)

	// Get key expression string using helper (handles version differences)
	keyStrData := C.view_string_data(&keystr)
	keyStrLen := C.view_string_len(&keystr)
	keyExprStr := C.GoStringN(keyStrData, C.int(keyStrLen))

	// Extract payload
	payload := C.z_sample_payload(sample)
	payloadLen := C.z_bytes_len(payload)

	var payloadData []byte
	if payloadLen > 0 {
		payloadData = make([]byte, payloadLen)
		// Use helper to read bytes safely
		C.read_bytes_to_buffer(payload, (*C.uint8_t)(unsafe.Pointer(&payloadData[0])), payloadLen)
	}

	s := Sample{
		KeyExpr:   KeyExpr(keyExprStr),
		Payload:   payloadData,
		Timestamp: time.Now(),
		Kind:      SampleKindPut,
	}

	// Call handler (in current goroutine - Zenoh manages threading)
	sub.handler(s)
}

