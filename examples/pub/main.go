// Example publisher for zenoh-go
//
// Usage:
//
//	go run ./examples/pub -e tcp/192.168.68.80:7447 -k reachy_mini/test
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	zenoh "github.com/teslashibe/zenoh-go"
)

func main() {
	endpoint := flag.String("e", "tcp/localhost:7447", "Zenoh endpoint")
	keyExpr := flag.String("k", "demo/example", "Key expression to publish to")
	interval := flag.Duration("i", time.Second, "Publish interval")
	flag.Parse()

	log.Printf("Connecting to %s...", *endpoint)

	session, err := zenoh.Open(zenoh.ClientConfig(*endpoint))
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close()

	info := session.Info()
	log.Printf("Session opened (CGO: %v)", info.UsingCGO)

	pub, err := session.Publisher(zenoh.KeyExpr(*keyExpr))
	if err != nil {
		log.Fatalf("Failed to declare publisher: %v", err)
	}

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	count := 0
	log.Printf("Publishing to '%s' every %v (Ctrl+C to stop)", *keyExpr, *interval)

	for {
		select {
		case <-sigChan:
			log.Println("Shutting down...")
			return
		case <-ticker.C:
			count++
			msg := fmt.Sprintf(`{"count": %d, "timestamp": "%s"}`, count, time.Now().Format(time.RFC3339))
			if err := pub.Put([]byte(msg)); err != nil {
				log.Printf("Put failed: %v", err)
			} else {
				log.Printf("Published: %s", msg)
			}
		}
	}
}









