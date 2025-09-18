package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"distributed-worker-system/pkg/coordinator"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Printf("🔗 Connected to NATS at %s", nats.DefaultURL)

	// Create coordinator instance
	coord := coordinator.NewCoordinator(nc, 8080)

	// Start results subscription in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := coord.SubscribeResults(ctx); err != nil {
			log.Printf("❌ Failed to subscribe to results: %v", err)
		}
	}()

	// Start coordinator HTTP server in a goroutine
	go func() {
		coord.StartHTTPServer()
	}()

	log.Printf("🎯 Coordinator started successfully!")
	log.Printf("📡 Coordinator API available at: http://localhost:8080")
	log.Printf("🚀 Submit requests at: POST http://localhost:8080/request")
	log.Printf("💡 Example request submission:")
	log.Printf("   curl -X POST http://localhost:8080/request \\")
	log.Printf("     -H 'Content-Type: application/json' \\")
	log.Printf("     -d '{\"query\":\"BTC/USD\"}'")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Printf("🛑 Shutting down coordinator...")
	cancel()
	coord.Close()
}
