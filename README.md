# zenoh-go

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Zenoh Version](https://img.shields.io/badge/zenoh--c-1.0.0-blue)](https://github.com/eclipse-zenoh/zenoh-c)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Alpha](https://img.shields.io/badge/status-alpha-orange)](https://github.com/teslashibe/zenoh-go/releases)

Go bindings for [Eclipse Zenoh](https://zenoh.io/) — a pub/sub/query protocol designed for robotics and IoT.

## Motivation

**Why zenoh-go?**

Eclipse Zenoh is increasingly adopted in robotics (ROS2, autonomous vehicles, drones) for its low-latency, peer-to-peer communication. However, Go developers have been left without native bindings — until now.

This project was born from a real need: **controlling the [Reachy Mini](https://www.pollen-robotics.com/reachy-mini/) robot from Go**. The robot's internal systems communicate via Zenoh at 100Hz+, and HTTP/WebSocket wrappers introduce unacceptable latency for real-time control loops.

| Communication Method | Latency | Suitable For |
|---------------------|---------|--------------|
| HTTP REST | 10-50ms | Configuration, status queries |
| WebSocket | 5-20ms | State streaming |
| **Zenoh (direct)** | **<1ms** | **Real-time motor control** |

With `zenoh-go`, you get direct access to Zenoh's pub/sub from Go — enabling sub-millisecond control loops, efficient joint position streaming, and true real-time robotics.

## Features

- **CGO Bindings** — Native performance via zenoh-c 1.0.0
- **Idiomatic Go API** — Channels, contexts, and error handling done right
- **Pub/Sub** — Publish and subscribe to key expressions with wildcard support
- **Query/Reply** — Request-response pattern for state queries
- **Mock Mode** — Full mock implementation for testing without zenoh-c
- **Thread-Safe** — Safe for concurrent use from multiple goroutines

## Installation

```bash
go get github.com/teslashibe/zenoh-go@v0.1.0-alpha
```

### Prerequisites

The CGO bindings require zenoh-c to be installed on your system:

**Ubuntu/Debian (ARM64):**
```bash
curl -L https://github.com/eclipse-zenoh/zenoh-c/releases/download/1.0.0/zenoh-c-1.0.0-aarch64-unknown-linux-gnu.deb -o zenoh-c.deb
sudo dpkg -i zenoh-c.deb
```

**Ubuntu/Debian (x86_64):**
```bash
curl -L https://github.com/eclipse-zenoh/zenoh-c/releases/download/1.0.0/zenoh-c-1.0.0-x86_64-unknown-linux-gnu.deb -o zenoh-c.deb
sudo dpkg -i zenoh-c.deb
```

**macOS (Homebrew):**
```bash
brew tap eclipse-zenoh/homebrew-zenoh
brew install zenoh-c
```

## Quick Start

### Publishing Data

```go
package main

import (
    "log"
    zenoh "github.com/teslashibe/zenoh-go"
)

func main() {
    // Connect to a Zenoh router
    session, err := zenoh.Open(zenoh.ClientConfig("tcp/192.168.1.100:7447"))
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // Create a publisher
    pub, err := session.Publisher("robot/command")
    if err != nil {
        log.Fatal(err)
    }
    defer pub.Close()

    // Publish motor commands at 100Hz
    for {
        pub.Put([]byte(`{"head_pose": [[1,0,0,0],[0,1,0,0],[0,0,1,0],[0,0,0,1]]}`))
        time.Sleep(10 * time.Millisecond)
    }
}
```

### Subscribing to Data

```go
package main

import (
    "log"
    zenoh "github.com/teslashibe/zenoh-go"
)

func main() {
    session, err := zenoh.Open(zenoh.ClientConfig("tcp/192.168.1.100:7447"))
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // Subscribe with wildcard support
    sub, err := session.Subscribe("robot/**", func(sample zenoh.Sample) {
        log.Printf("[%s] %s", sample.KeyExpr, sample.Payload)
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Close()

    // Keep running
    select {}
}
```

### Real-World Example: Robot Control

```go
// Connect to Reachy Mini robot
session, _ := zenoh.Open(zenoh.ClientConfig("tcp/192.168.68.83:7447"))
defer session.Close()

// Subscribe to joint positions (100Hz feedback)
session.Subscribe("reachy_mini/joint_positions", func(s zenoh.Sample) {
    var joints JointPositions
    json.Unmarshal(s.Payload, &joints)
    log.Printf("Head: roll=%.2f pitch=%.2f yaw=%.2f", 
        joints.Head[0], joints.Head[1], joints.Head[2])
})

// Publish motor commands
pub, _ := session.Publisher("reachy_mini/command")
pub.Put([]byte(`{"head_pose": [...], "antennas": [0.5, -0.5]}`))
```

## API Reference

| Function | Description |
|----------|-------------|
| `Open(Config)` | Create a new Zenoh session |
| `session.Publisher(KeyExpr)` | Declare a publisher for a key expression |
| `session.Subscribe(KeyExpr, Handler)` | Subscribe to a key expression (supports `*` and `**` wildcards) |
| `session.Get(ctx, KeyExpr)` | Query for samples (request/reply pattern) |
| `session.Close()` | Close session and release resources |
| `session.Info()` | Get session metadata (ID, mode, endpoints) |

### Configuration

```go
// Client mode - connect to specific endpoints
cfg := zenoh.ClientConfig("tcp/192.168.1.100:7447")

// Peer mode - use multicast discovery
cfg := zenoh.PeerConfig()

// Custom configuration
cfg := zenoh.Config{
    Mode:      zenoh.ModeClient,
    Endpoints: []string{"tcp/host1:7447", "tcp/host2:7447"},
}
```

## Mock Mode (Testing)

When `CGO_ENABLED=0`, a mock implementation is automatically used. This allows you to:
- Run tests without installing zenoh-c
- Test in CI/CD pipelines
- Develop on systems without Zenoh

```bash
# Run tests with mock (no zenoh-c required)
CGO_ENABLED=0 go test ./...

# Run tests with real CGO bindings
CGO_ENABLED=1 go test ./...
```

## Development

```bash
# Clone the repository
git clone https://github.com/teslashibe/zenoh-go.git
cd zenoh-go

# Test with mock implementation
make test-mock

# Test with CGO (requires zenoh-c)
make test-cgo

# Build examples
make examples

# Run publisher example
./bin/pub

# Run subscriber example (in another terminal)
./bin/sub
```

## Compatibility

| zenoh-go | zenoh-c | Go |
|----------|---------|-----|
| v0.1.0-alpha | 1.0.0 | 1.21+ |

## Roadmap

- [ ] Queryable (reply to queries)
- [ ] Liveliness tokens
- [ ] Attachment support
- [ ] SHM (shared memory) transport
- [ ] More configuration options

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Eclipse Zenoh](https://zenoh.io/) — The underlying pub/sub protocol
- [Pollen Robotics](https://www.pollen-robotics.com/) — Reachy Mini robot that inspired this project

