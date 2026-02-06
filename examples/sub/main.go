// Example subscriber for zenoh-go
//
// Usage:
//
//	go run ./examples/sub -e tcp/192.168.68.80:7447 -k reachy_mini/**
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	zenoh "github.com/evaeverywhere/zenoh-go"
)

func main() {
	endpoint := flag.String("e", "tcp/localhost:7447", "Zenoh endpoint")
	keyExpr := flag.String("k", "demo/**", "Key expression to subscribe to")
	flag.Parse()

	log.Printf("Connecting to %s...", *endpoint)

	session, err := zenoh.Open(zenoh.ClientConfig(*endpoint))
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close()

	info := session.Info()
	log.Printf("Session opened (CGO: %v)", info.UsingCGO)

	log.Printf("Subscribing to '%s'...", *keyExpr)

	sub, err := session.Subscribe(zenoh.KeyExpr(*keyExpr), func(s zenoh.Sample) {
		log.Printf("[%s] %s: %s", s.Kind, s.KeyExpr, s.Payload)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Close()

	log.Println("Waiting for samples (Ctrl+C to stop)...")

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}









