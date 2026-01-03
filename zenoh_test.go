package zenoh

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestOpenSession(t *testing.T) {
	session, err := Open(ClientConfig("tcp/localhost:7447"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer session.Close()

	info := session.Info()
	if info.Mode != ModeClient {
		t.Errorf("Expected mode %q, got %q", ModeClient, info.Mode)
	}
	if info.UsingCGO {
		t.Error("Expected UsingCGO=false for mock")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid client config",
			config:  ClientConfig("tcp/localhost:7447"),
			wantErr: false,
		},
		{
			name:    "valid peer config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "client without endpoints",
			config:  Config{Mode: ModeClient, ConnectTimeout: time.Second},
			wantErr: true,
		},
		{
			name:    "invalid mode",
			config:  Config{Mode: "invalid", ConnectTimeout: time.Second},
			wantErr: true,
		},
		{
			name:    "zero timeout",
			config:  Config{Mode: ModePeer, ConnectTimeout: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPubSub(t *testing.T) {
	session, err := Open(DefaultConfig())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer session.Close()

	// Subscribe first
	var received []Sample
	var mu sync.Mutex
	done := make(chan struct{})
	var closeOnce sync.Once

	sub, err := session.Subscribe("test/topic", func(s Sample) {
		mu.Lock()
		received = append(received, s)
		count := len(received)
		mu.Unlock()
		if count >= 3 {
			closeOnce.Do(func() { close(done) })
		}
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Close()

	// Publish
	pub, err := session.Publisher("test/topic")
	if err != nil {
		t.Fatalf("Publisher failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := pub.Put([]byte("hello")); err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	}

	// Wait for messages
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for messages")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(received))
	}
}

func TestWildcardSubscription(t *testing.T) {
	session, err := Open(DefaultConfig())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer session.Close()

	var received []Sample
	var mu sync.Mutex

	// Subscribe with wildcard
	sub, err := session.Subscribe("test/*", func(s Sample) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Close()

	// Publish to different topics
	pub1, _ := session.Publisher("test/a")
	pub2, _ := session.Publisher("test/b")
	pub3, _ := session.Publisher("other/c")

	pub1.Put([]byte("a"))
	pub2.Put([]byte("b"))
	pub3.Put([]byte("c")) // Should not match

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Errorf("Expected 2 messages (wildcard match), got %d", len(received))
	}
}

func TestDoubleStarWildcard(t *testing.T) {
	session, err := Open(DefaultConfig())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer session.Close()

	var received []Sample
	var mu sync.Mutex

	// Subscribe with ** wildcard
	sub, err := session.Subscribe("reachy/**", func(s Sample) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Close()

	// Publish to nested topics
	pub1, _ := session.Publisher("reachy/a")
	pub2, _ := session.Publisher("reachy/a/b")
	pub3, _ := session.Publisher("reachy/a/b/c")
	pub4, _ := session.Publisher("other/x")

	pub1.Put([]byte("1"))
	pub2.Put([]byte("2"))
	pub3.Put([]byte("3"))
	pub4.Put([]byte("4")) // Should not match

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Errorf("Expected 3 messages (** match), got %d", len(received))
	}
}

func TestGet(t *testing.T) {
	session, err := Open(DefaultConfig())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer session.Close()

	// Publish some data
	pub, _ := session.Publisher("data/key1")
	pub.Put([]byte("value1"))

	pub2, _ := session.Publisher("data/key2")
	pub2.Put([]byte("value2"))

	// Query
	ctx := context.Background()
	samples, err := session.Get(ctx, "data/key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(samples) != 1 {
		t.Errorf("Expected 1 sample, got %d", len(samples))
	}
}

func TestSessionClose(t *testing.T) {
	session, err := Open(DefaultConfig())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	session.Close()

	// Operations should fail after close
	_, err = session.Publisher("test")
	if err != ErrSessionClosed {
		t.Errorf("Expected ErrSessionClosed, got %v", err)
	}

	_, err = session.Subscribe("test", func(s Sample) {})
	if err != ErrSessionClosed {
		t.Errorf("Expected ErrSessionClosed, got %v", err)
	}
}

func TestKeyExprMatch(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		{"a/b", "a/b", true},
		{"a/b", "a/c", false},
		{"a/*", "a/b", true},
		{"a/*", "a/b/c", false},
		{"a/**", "a/b", true},
		{"a/**", "a/b/c", true},
		{"a/**", "a/b/c/d", true},
		{"**", "anything", true},
		{"**", "a/b/c", true},
		{"a/*/c", "a/b/c", true},
		{"a/*/c", "a/b/d", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			got := matchKeyExpr(KeyExpr(tt.pattern), KeyExpr(tt.subject))
			if got != tt.want {
				t.Errorf("matchKeyExpr(%q, %q) = %v, want %v", tt.pattern, tt.subject, got, tt.want)
			}
		})
	}
}

