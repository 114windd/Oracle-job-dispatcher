package main

import (
	"log"
	"net/http"
	"time"

	"distributed-worker-system/pkg/client"
)

func main() {
	// Demo queries
	queries := []string{
		"BTC/USD",
		"ETH/USD",
		"SOL/USD",
		"MATIC/USD",
		"BTC/USD", // Repeat to show variance
	}

	log.Printf("🎬 Starting Distributed Worker System Demo")
	log.Printf("📝 Make sure to start the coordinator and workers first:")
	log.Printf("   go run cmd/coordinator/main.go")
	log.Printf("   go run cmd/worker/main.go -port=8081")
	log.Printf("   go run cmd/worker/main.go -port=8082")
	log.Printf("   go run cmd/worker/main.go -port=8083")

	// Check if coordinator is running
	coordinatorURL := "http://localhost:8080"
	if !isCoordinatorRunning(coordinatorURL) {
		log.Printf("❌ Coordinator is not running on %s", coordinatorURL)
		log.Printf("💡 Please start the coordinator first:")
		log.Printf("   go run cmd/coordinator/main.go")
		return
	}

	log.Printf("✅ Coordinator is running!")

	// Wait a moment for workers to register
	log.Printf("⏳ Waiting 2 seconds for workers to register...")
	time.Sleep(2 * time.Second)

	// Run client simulation
	client.SimulateClient(coordinatorURL, queries)
}

// isCoordinatorRunning checks if the coordinator is running
func isCoordinatorRunning(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
