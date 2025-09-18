package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"distributed-worker-system/pkg/models"
	"distributed-worker-system/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

// Coordinator manages workers, tasks, and aggregation using NATS
type Coordinator struct {
	nc          *nats.Conn
	port        int
	pendingReqs map[string]chan models.WorkerResult
	pendingMux  sync.RWMutex
	resultsSub  *nats.Subscription
}

// NewCoordinator initializes coordinator with NATS connection
func NewCoordinator(nc *nats.Conn, port int) *Coordinator {
	return &Coordinator{
		nc:          nc,
		port:        port,
		pendingReqs: make(map[string]chan models.WorkerResult),
	}
}

// PublishTask sends an oracle request into NATS
func (c *Coordinator) PublishTask(req models.OracleRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	if err := c.nc.Publish("oracle.tasks", reqBytes); err != nil {
		return fmt.Errorf("failed to publish task: %v", err)
	}

	log.Printf("📤 Published task %s to oracle.tasks", req.ID)
	return nil
}

// SubscribeResults listens on NATS for worker results
func (c *Coordinator) SubscribeResults(ctx context.Context) error {
	sub, err := c.nc.Subscribe("oracle.results", func(msg *nats.Msg) {
		var result models.WorkerResult
		if err := json.Unmarshal(msg.Data, &result); err != nil {
			log.Printf("❌ Failed to unmarshal worker result: %v", err)
			return
		}

		log.Printf("📥 Received result from worker %s for request %s", result.WorkerID, result.RequestID)
		c.handleWorkerResult(result)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to oracle.results: %v", err)
	}

	c.resultsSub = sub
	log.Printf("👂 Subscribed to oracle.results")

	// Keep subscription alive
	<-ctx.Done()
	return sub.Unsubscribe()
}

// handleWorkerResult processes incoming worker results
func (c *Coordinator) handleWorkerResult(result models.WorkerResult) {
	c.pendingMux.RLock()
	ch, exists := c.pendingReqs[result.RequestID]
	c.pendingMux.RUnlock()

	if !exists {
		log.Printf("⚠️  Received result for unknown request %s", result.RequestID)
		return
	}

	// Send result to waiting goroutine
	select {
	case ch <- result:
	default:
		log.Printf("⚠️  Channel full for request %s", result.RequestID)
	}
}

// SubmitRequest submits an oracle request and waits for results (with timeout)
func (c *Coordinator) SubmitRequest(ctx context.Context, req models.OracleRequest) models.OracleResult {
	log.Printf("🚀 Processing request %s: %s", req.ID, req.Query)

	// Create channel for this request
	resultChan := make(chan models.WorkerResult, 10) // Buffer for multiple workers
	c.pendingMux.Lock()
	c.pendingReqs[req.ID] = resultChan
	c.pendingMux.Unlock()

	// Clean up when done
	defer func() {
		c.pendingMux.Lock()
		delete(c.pendingReqs, req.ID)
		c.pendingMux.Unlock()
		close(resultChan)
	}()

	// Publish task to NATS
	if err := c.PublishTask(req); err != nil {
		return models.OracleResult{
			RequestID:       req.ID,
			FinalValue:      0,
			WorkerResponses: []models.WorkerResult{},
			ReliabilityNote: fmt.Sprintf("Failed to publish task: %v", err),
		}
	}

	// Collect results with timeout
	var workerResults []models.WorkerResult
	timeout := time.After(5 * time.Second)

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// Channel closed, return what we have
				return c.aggregateResults(req.ID, workerResults)
			}
			workerResults = append(workerResults, result)
		case <-timeout:
			log.Printf("⏰ Timeout waiting for worker responses for request %s", req.ID)
			return c.aggregateResults(req.ID, workerResults)
		case <-ctx.Done():
			log.Printf("❌ Context cancelled for request %s", req.ID)
			return c.aggregateResults(req.ID, workerResults)
		}
	}
}

// aggregateResults aggregates worker results and returns final result
func (c *Coordinator) aggregateResults(requestID string, results []models.WorkerResult) models.OracleResult {
	// Aggregate results using average
	finalValue := AggregateResults(results, "average")

	// Calculate reliability note
	reliabilityNote := c.calculateReliabilityNote(results)

	result := models.OracleResult{
		RequestID:       requestID,
		FinalValue:      finalValue,
		WorkerResponses: results,
		ReliabilityNote: reliabilityNote,
	}

	utils.LogOracleResult(result)
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

// StartHTTPServer starts the coordinator HTTP server with middleware
func (c *Coordinator) StartHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Initialize middleware
	rateLimiter := NewRateLimitMiddleware(10.0, 20) // 10 requests/sec, burst of 20
	rateLimiter.StartCleanup()

	// Apply middleware
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			WriteJSONError(c.Writer, "internal server error", 500, err)
		} else {
			WriteJSONError(c.Writer, "internal server error", 500, "an unexpected error occurred")
		}
		c.Abort()
	}))

	// Add custom middleware for rate limiting and logging
	r.Use(func(c *gin.Context) {
		// Rate limiting
		rateLimiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})

	// Register routes
	r.GET("/health", c.handleHealth)
	r.POST("/request", c.handleRequest)

	log.Printf("🌐 Coordinator server starting on port %d with rate limiting (10 req/sec)", c.port)
	if err := r.Run(fmt.Sprintf(":%d", c.port)); err != nil {
		log.Fatalf("Failed to start coordinator server: %v", err)
	}
}

// handleRequest handles oracle request submissions
func (c *Coordinator) handleRequest(ctx *gin.Context) {
	var req models.OracleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		WriteJSONError(ctx.Writer, "invalid request", 400, err.Error())
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

	// Check if we got any results
	if len(result.WorkerResponses) == 0 {
		WriteJSONError(ctx.Writer, "no workers available", 503, "no workers responded to the request")
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// handleHealth handles health check requests
func (c *Coordinator) handleHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"port":   c.port,
		"nats":   c.nc.IsConnected(),
	})
}

// Close cleans up NATS connection
func (c *Coordinator) Close() error {
	if c.resultsSub != nil {
		c.resultsSub.Unsubscribe()
	}
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}
