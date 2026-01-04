# zenoh-go

Go bindings for [Eclipse Zenoh](https://zenoh.io/).

> ⚠️ **Work in Progress** - This library is under active development.

## Overview

This package provides CGO bindings to the zenoh-c library, enabling Go applications
to use Zenoh's pub/sub, query, and storage capabilities.

## Installation

```bash
go get github.com/teslashibe/zenoh-go
```

### Prerequisites

The CGO bindings require zenoh-c to be installed:

**Ubuntu/Debian (ARM64):**
```bash
curl -L https://github.com/eclipse-zenoh/zenoh-c/releases/download/1.0.0/zenoh-c-1.0.0-aarch64-unknown-linux-gnu.deb -o zenoh-c.deb
sudo dpkg -i zenoh-c.deb
```

**macOS:**
```bash
brew tap eclipse-zenoh/homebrew-zenoh
brew install zenoh-c
```

## Usage

### Basic Pub/Sub

```go
package main

import (
    "log"
    zenoh "github.com/teslashibe/zenoh-go"
)

func main() {
    // Connect to Zenoh router
    session, err := zenoh.Open(zenoh.ClientConfig("tcp/192.168.68.80:7447"))
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // Subscribe
    session.Subscribe("reachy_mini/joint_positions", func(s zenoh.Sample) {
        log.Printf("Received: %s", s.Payload)
    })

    // Publish
    pub, _ := session.Publisher("reachy_mini/command")
    pub.Put([]byte(`{"head": [0, 0, 0, 0.5]}`))
}
```

### Mock Mode (Testing)

When CGO is disabled, a mock implementation is used:

```bash
CGO_ENABLED=0 go test ./...
```

## API

| Function | Description |
|----------|-------------|
| `Open(Config)` | Create a new session |
| `session.Publisher(KeyExpr)` | Declare a publisher |
| `session.Subscribe(KeyExpr, Handler)` | Subscribe to key expression |
| `session.Get(ctx, KeyExpr)` | Query for samples |
| `session.Close()` | Close session |

## Build Tags

| Mode | Description |
|------|-------------|
| `CGO_ENABLED=1` (default) | Use native zenoh-c bindings |
| `CGO_ENABLED=0` | Use mock implementation for testing |

## Development

```bash
# Test with mock (no zenoh-c required)
make test-mock

# Test with CGO (requires zenoh-c)
make test-cgo

# Build examples
make examples
```

## License

Apache 2.0





