package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"distributed-worker-system/pkg/worker"

	"github.com/nats-io/nats.go"
)

func main() {
	// Parse command line flags
	var port = flag.Int("port", 8081, "Port for the worker (used for identification)")
	flag.Parse()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Printf("🔗 Connected to NATS at %s", nats.DefaultURL)

	// Create worker instance
	w := worker.NewWorker(*port)

	log.Printf("🔧 Worker %s started successfully!", w.GetID())
	log.Printf("👂 Subscribing to oracle.tasks...")

	// Subscribe to tasks and process them
	if err := w.SubscribeTasks(nc); err != nil {
		log.Fatalf("Failed to subscribe to tasks: %v", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Printf("🛑 Shutting down worker %s...", w.GetID())
	w.Close()
}
