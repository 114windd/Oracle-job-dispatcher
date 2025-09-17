package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"distributed-worker-system/pkg/models"
	"distributed-worker-system/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Coordinator manages workers, tasks, and aggregation
type Coordinator struct {
	workers    map[string]models.WorkerInfo
	workersMux sync.RWMutex
	httpClient *http.Client
	aggregator string
	port       int
}

// NewCoordinator initializes coordinator with HTTP client and worker registry
func NewCoordinator(port int) *Coordinator {
	return &Coordinator{
		workers: make(map[string]models.WorkerInfo),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		aggregator: "average", // default strategy
		port:       port,
	}
}

// RegisterWorker registers a new worker with the coordinator
func (c *Coordinator) RegisterWorker(workerInfo models.WorkerInfo) error {
	c.workersMux.Lock()
	defer c.workersMux.Unlock()

	workerInfo.LastSeen = time.Now()
	c.workers[workerInfo.ID] = workerInfo

	log.Printf("📝 Worker %s registered at %s", workerInfo.ID, workerInfo.Endpoint)
	return nil
}

// SubmitRequest submits an oracle request and waits for results (with timeout)
func (c *Coordinator) SubmitRequest(ctx context.Context, req models.OracleRequest) models.OracleResult {
	log.Printf("🚀 Processing request %s: %s", req.ID, req.Query)

	// Dispatch to all available workers
	workerResults := c.dispatchToWorkers(ctx, req)

	// Aggregate results
	finalValue := AggregateResults(workerResults, c.aggregator)

	// Calculate reliability note
	reliabilityNote := c.calculateReliabilityNote(workerResults)

	result := models.OracleResult{
		RequestID:       req.ID,
		FinalValue:      finalValue,
		WorkerResponses: workerResults,
		ReliabilityNote: reliabilityNote,
	}

	utils.LogOracleResult(result)
	return result
}

// dispatchToWorkers sends request to all registered workers concurrently
func (c *Coordinator) dispatchToWorkers(ctx context.Context, req models.OracleRequest) []models.WorkerResult {
	c.workersMux.RLock()
	workerList := make([]models.WorkerInfo, 0, len(c.workers))
	for _, worker := range c.workers {
		workerList = append(workerList, worker)
	}
	c.workersMux.RUnlock()

	if len(workerList) == 0 {
		log.Printf("⚠️  No workers available for request %s", req.ID)
		return []models.WorkerResult{}
	}

	// Create channels for concurrent processing
	results := make(chan models.WorkerResult, len(workerList))
	var wg sync.WaitGroup

	// Send request to each worker concurrently
	for _, worker := range workerList {
		wg.Add(1)
		go func(w models.WorkerInfo) {
			defer wg.Done()
			result := c.sendTaskToWorker(ctx, req, w)
			results <- result
		}(worker)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with timeout
	var workerResults []models.WorkerResult
	timeout := time.After(3 * time.Second)

	for {
		select {
		case result, ok := <-results:
			if !ok {
				return workerResults
			}
			workerResults = append(workerResults, result)
		case <-timeout:
			log.Printf("⏰ Timeout waiting for worker responses for request %s", req.ID)
			return workerResults
		case <-ctx.Done():
			log.Printf("❌ Context cancelled for request %s", req.ID)
			return workerResults
		}
	}
}

// sendTaskToWorker sends a task to a specific worker via HTTP
func (c *Coordinator) sendTaskToWorker(ctx context.Context, req models.OracleRequest, worker models.WorkerInfo) models.WorkerResult {
	startTime := time.Now()

	// Prepare request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return models.WorkerResult{
			WorkerID:     worker.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          fmt.Sprintf("failed to marshal request: %v", err),
			ResponseTime: time.Since(startTime),
			Reliable:     false,
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", worker.Endpoint+"/task", bytes.NewBuffer(reqBody))
	if err != nil {
		return models.WorkerResult{
			WorkerID:     worker.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          fmt.Sprintf("failed to create HTTP request: %v", err),
			ResponseTime: time.Since(startTime),
			Reliable:     false,
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return models.WorkerResult{
			WorkerID:     worker.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          fmt.Sprintf("HTTP request failed: %v", err),
			ResponseTime: time.Since(startTime),
			Reliable:     false,
		}
	}
	defer resp.Body.Close()

	// Parse response using Gin's JSON utilities
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.WorkerResult{
			WorkerID:     worker.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          fmt.Sprintf("failed to read response body: %v", err),
			ResponseTime: time.Since(startTime),
			Reliable:     false,
		}
	}

	var result models.WorkerResult
	if err := json.Unmarshal(body, &result); err != nil {
		return models.WorkerResult{
			WorkerID:     worker.ID,
			RequestID:    req.ID,
			Value:        0,
			Err:          fmt.Sprintf("failed to decode response: %v", err),
			ResponseTime: time.Since(startTime),
			Reliable:     false,
		}
	}

	// Update response time
	result.ResponseTime = time.Since(startTime)
	result.WorkerID = worker.ID
	result.RequestID = req.ID

	return result
}

// calculateReliabilityNote calculates a reliability note based on worker responses
func (c *Coordinator) calculateReliabilityNote(results []models.WorkerResult) string {
	if len(results) == 0 {
		return "No workers responded"
	}

	successCount := 0
	for _, result := range results {
		if result.Err == "" {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(len(results))

	if successRate >= 0.8 {
		return "All workers responded successfully"
	} else if successRate >= 0.5 {
		return fmt.Sprintf("Partial response: %d/%d workers succeeded", successCount, len(results))
	} else {
		return fmt.Sprintf("Low reliability: only %d/%d workers succeeded", successCount, len(results))
	}
}

// StartHTTPServer starts the coordinator HTTP server
func (c *Coordinator) StartHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Register routes
	r.GET("/health", c.handleHealth)
	r.POST("/register", c.handleRegister)
	r.POST("/request", c.handleRequest)

	log.Printf("🌐 Coordinator server starting on port %d", c.port)
	if err := r.Run(fmt.Sprintf(":%d", c.port)); err != nil {
		log.Fatalf("Failed to start coordinator server: %v", err)
	}
}

// handleRegister handles worker registration requests
func (c *Coordinator) handleRegister(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workerInfo := models.WorkerInfo{
		ID:       req.ID,
		Endpoint: req.Endpoint,
		LastSeen: time.Now(),
		Reliable: true,
	}

	if err := c.RegisterWorker(workerInfo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.RegisterResponse{
		Status:  "registered",
		Message: fmt.Sprintf("Worker %s successfully registered", req.ID),
	}

	ctx.JSON(http.StatusOK, response)
}

// handleRequest handles oracle request submissions
func (c *Coordinator) handleRequest(ctx *gin.Context) {
	var req models.OracleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate request ID if not provided
	if req.ID == "" {
		req.ID = utils.GenerateRequestID()
	}

	// Process request with timeout
	requestCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := c.SubmitRequest(requestCtx, req)
	ctx.JSON(http.StatusOK, result)
}

// GetWorkerCount returns the number of registered workers
func (c *Coordinator) GetWorkerCount() int {
	c.workersMux.RLock()
	defer c.workersMux.RUnlock()
	return len(c.workers)
}

// GetWorkers returns a copy of all registered workers
func (c *Coordinator) GetWorkers() []models.WorkerInfo {
	c.workersMux.RLock()
	defer c.workersMux.RUnlock()

	workers := make([]models.WorkerInfo, 0, len(c.workers))
	for _, worker := range c.workers {
		workers = append(workers, worker)
	}
	return workers
}

// handleHealth handles health check requests
func (c *Coordinator) handleHealth(ctx *gin.Context) {
	c.workersMux.RLock()
	workerCount := len(c.workers)
	c.workersMux.RUnlock()

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"workers": workerCount,
		"port":    c.port,
	})
}
