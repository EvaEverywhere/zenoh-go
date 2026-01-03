package zenoh

import (
	"errors"
	"fmt"
	"time"
)

// Config holds Zenoh session configuration.
type Config struct {
	// Mode is either "peer" or "client".
	// Client mode requires at least one endpoint.
	// Peer mode can use multicast discovery.
	Mode string

	// Endpoints to connect to (e.g., "tcp/192.168.68.80:7447").
	// Required for client mode.
	Endpoints []string

	// ListenEndpoints for peer mode (optional).
	// E.g., "tcp/0.0.0.0:7447"
	ListenEndpoints []string

	// ConnectTimeout for connection attempts.
	// Default: 5 seconds
	ConnectTimeout time.Duration
}

// DefaultConfig returns a config for local peer mode.
func DefaultConfig() Config {
	return Config{
		Mode:           ModePeer,
		ConnectTimeout: 5 * time.Second,
	}
}

// ClientConfig returns a config for connecting to specific endpoints.
func ClientConfig(endpoints ...string) Config {
	return Config{
		Mode:           ModeClient,
		Endpoints:      endpoints,
		ConnectTimeout: 5 * time.Second,
	}
}

// PeerConfig returns a config for peer mode with optional listen endpoints.
func PeerConfig(listenEndpoints ...string) Config {
	return Config{
		Mode:            ModePeer,
		ListenEndpoints: listenEndpoints,
		ConnectTimeout:  5 * time.Second,
	}
}

// Validate checks the configuration for errors.
func (c Config) Validate() error {
	if c.Mode == ModeClient && len(c.Endpoints) == 0 {
		return errors.New("client mode requires at least one endpoint")
	}
	if c.Mode != ModePeer && c.Mode != ModeClient {
		return fmt.Errorf("invalid mode: %s (must be %q or %q)", c.Mode, ModePeer, ModeClient)
	}
	if c.ConnectTimeout <= 0 {
		return errors.New("connect timeout must be positive")
	}
	return nil
}

// WithTimeout returns a copy of the config with the specified timeout.
func (c Config) WithTimeout(timeout time.Duration) Config {
	c.ConnectTimeout = timeout
	return c
}

