package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"distributed-worker-system/pkg/models"
	"distributed-worker-system/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Worker represents a worker that processes oracle tasks
type Worker struct {
	ID        string
	Port      int
	Reliable  bool
	baseValue float64
}

// NewWorker creates a new worker instance
func NewWorker(port int) *Worker {
	return &Worker{
		ID:        utils.GenerateWorkerID(),
		Port:      port,
		Reliable:  true,
		baseValue: 50000.0, // Base value for simulationE
	}
}

// StartWorkerServer starts the worker HTTP server
func (w *Worker) StartWorkerServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Register routes
	r.POST("/task", w.handleTask)
	r.GET("/health", w.handleHealth)

	log.Printf("🔧 Worker %s starting on port %d", w.ID, w.Port)
	if err := r.Run(fmt.Sprintf(":%d", w.Port)); err != nil {
		log.Fatalf("Failed to start worker server: %v", err)
	}
}

// handleTask handles task processing requests
func (w *Worker) handleTask(ctx *gin.Context) {
	var req models.OracleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📋 Worker %s processing task %s: %s", w.ID, req.ID, req.Query)

	// Process the task
	result := w.processTask(req)

	ctx.JSON(http.StatusOK, result)
}

// handleHealth handles health check requests
func (w *Worker) handleHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"worker_id": w.ID,
		"reliable":  w.Reliable,
	})
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
			Reliable:     w.Reliable,
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
		Reliable:     w.Reliable,
	}
}

// simulateResponse generates a random value with variance based on query
func (w *Worker) simulateResponse(query string) float64 {
	// Base value varies by query type
	baseValue := w.baseValue

	// Adjust base value based on query
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

// RegisterWithCoordinator registers this worker with the coordinator using Gin-style approach
func (w *Worker) RegisterWithCoordinator(coordinatorURL string) error {
	registerReq := models.RegisterRequest{
		ID:       w.ID,
		Endpoint: fmt.Sprintf("http://localhost:%d", w.Port),
	}

	// Marshal request using Gin's JSON utilities
	reqBody, err := json.Marshal(registerReq)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", coordinatorURL+"/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register with coordinator: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coordinator returned status %d", resp.StatusCode)
	}

	log.Printf("✅ Worker %s registered with coordinator at %s", w.ID, coordinatorURL)
	return nil
}

// SetReliability sets the worker's reliability status
func (w *Worker) SetReliability(reliable bool) {
	w.Reliable = reliable
}

// GetID returns the worker's ID
func (w *Worker) GetID() string {
	return w.ID
}
