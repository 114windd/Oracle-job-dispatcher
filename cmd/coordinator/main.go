package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"distributed-worker-system/pkg/coordinator"
)

func main() {
	// Create coordinator instance
	coord := coordinator.NewCoordinator(8080)

	// Start coordinator server in a goroutine
	go func() {
		coord.StartHTTPServer()
	}()

	log.Printf("🎯 Coordinator started successfully!")
	log.Printf("📡 Coordinator API available at: http://localhost:8080")
	log.Printf("📝 Register workers at: POST http://localhost:8080/register")
	log.Printf("🚀 Submit requests at: POST http://localhost:8080/request")
	log.Printf("💡 Example worker registration:")
	log.Printf("   curl -X POST http://localhost:8080/register \\")
	log.Printf("     -H 'Content-Type: application/json' \\")
	log.Printf("     -d '{\"id\":\"worker-1\",\"endpoint\":\"http://localhost:8081\"}'")
	log.Printf("💡 Example request submission:")
	log.Printf("   curl -X POST http://localhost:8080/request \\")
	log.Printf("     -H 'Content-Type: application/json' \\")
	log.Printf("     -d '{\"query\":\"BTC/USD\"}'")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Printf("🛑 Shutting down coordinator...")
}
