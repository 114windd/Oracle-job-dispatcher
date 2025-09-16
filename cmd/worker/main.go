package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"distributed-worker-system/pkg/worker"
)

func main() {
	// Parse command line flags
	var port = flag.Int("port", 8081, "Port for the worker server")
	var coordinatorURL = flag.String("coordinator", "http://localhost:8080", "Coordinator URL")
	var register = flag.Bool("register", true, "Whether to register with coordinator")
	flag.Parse()

	// Create worker instance
	w := worker.NewWorker(*port)

	// Register with coordinator if requested
	if *register {
		maxRetries := 3
		registered := false
		for i := 0; i < maxRetries; i++ {
			if err := w.RegisterWithCoordinator(*coordinatorURL); err != nil {
				log.Printf("⚠️  Registration attempt %d failed: %v", i+1, err)
				if i < maxRetries-1 {
					time.Sleep(1 * time.Second)
				}
			} else {
				registered = true
				break
			}
		}
		if !registered {
			log.Printf("⚠️  Failed to register with coordinator after %d attempts. Worker will run independently.", maxRetries)
		}
	}

	log.Printf("🔧 Worker %s started successfully!", w.GetID())
	log.Printf("📡 Worker API available at: http://localhost:%d", *port)
	log.Printf("📋 Process tasks at: POST http://localhost:%d/task", *port)
	log.Printf("❤️  Health check at: GET http://localhost:%d/health", *port)

	// Start worker server in a goroutine
	go func() {
		w.StartWorkerServer()
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Printf("🛑 Shutting down worker %s...", w.GetID())
}
