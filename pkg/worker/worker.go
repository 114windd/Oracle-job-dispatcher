package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"distributed-worker-system/pkg/models"
	"distributed-worker-system/pkg/utils"

	"github.com/nats-io/nats.go"
)

// Worker represents a worker that processes oracle tasks via NATS
type Worker struct {
	ID   string
	Port int
	nc   *nats.Conn
}

// NewWorker creates a new worker instance
func NewWorker(port int) *Worker {
	return &Worker{
		ID:   utils.GenerateWorkerID(),
		Port: port,
	}
}

// SubscribeTasks listens for new tasks and processes them
func (w *Worker) SubscribeTasks(nc *nats.Conn) error {
	w.nc = nc

	// Subscribe to oracle.tasks subject
	sub, err := nc.Subscribe("oracle.tasks", func(msg *nats.Msg) {
		var req models.OracleRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			log.Printf("❌ Worker %s failed to unmarshal task: %v", w.ID, err)
			return
		}

		log.Printf("📋 Worker %s processing task %s: %s", w.ID, req.ID, req.Query)

		// Process the task
		result := w.processTask(req)

		// Publish result back to oracle.results
		if err := w.publishResult(result); err != nil {
			log.Printf("❌ Worker %s failed to publish result: %v", w.ID, err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to oracle.tasks: %v", err)
	}

	log.Printf("👂 Worker %s subscribed to oracle.tasks", w.ID)

	// Keep subscription alive
	<-context.Background().Done()
	return sub.Unsubscribe()
}

// publishResult publishes worker result to oracle.results subject
func (w *Worker) publishResult(result models.WorkerResult) error {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	if err := w.nc.Publish("oracle.results", resultBytes); err != nil {
		return fmt.Errorf("failed to publish result: %v", err)
	}

	log.Printf("📤 Worker %s published result for request %s", w.ID, result.RequestID)
	return nil
}

// processTask simulates fetching oracle data with latency/failure
func (w *Worker) processTask(req models.OracleRequest) models.WorkerResult {
	startTime := time.Now()

	// Simulate random delay (100ms to 2s)
	delay := time.Duration(rand.Intn(1900)+100) * time.Millisecond
	time.Sleep(delay)

	// Simulate occasional failures (10% chance)
	if rand.Float64() < 0.1 {
		return models.WorkerResult{
			WorkerID:     w.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          "simulated worker failure",
			ResponseTime: time.Since(startTime),
		}
	}

	// Simulate response with slight variance
	value := w.simulateResponse(req.Query)

	return models.WorkerResult{
		WorkerID:     w.ID,
		RequestID:    req.ID,
		Value:        value,
		Err:          "",
		ResponseTime: time.Since(startTime),
	}
}

// simulateResponse generates a random value with variance based on query
func (w *Worker) simulateResponse(query string) float64 {
	// Base value varies by query type
	var baseValue float64
	switch query {
	case "BTC/USD":
		baseValue = 42000.0
	case "ETH/USD":
		baseValue = 2500.0
	case "SOL/USD":
		baseValue = 100.0
	case "MATIC/USD":
		baseValue = 0.8
	default:
		baseValue = 1000.0
	}

	// Add random variance (±5%)
	variance := (rand.Float64() - 0.5) * 0.1 * baseValue
	return baseValue + variance
}

// GetID returns the worker's ID
func (w *Worker) GetID() string {
	return w.ID
}

// Close cleans up NATS connection
func (w *Worker) Close() error {
	if w.nc != nil {
		w.nc.Close()
	}
	return nil
}
