📄 PRD: Distributed Worker System – Simulating Oracle Requests

Project Overview

The goal of this project is to build a distributed worker system in Go that simulates oracle requests. A central Coordinator will accept incoming requests, dispatch them to multiple Workers, and then collect and aggregate the results into a final answer. This will mimic the behavior of blockchain oracle networks, where data is retrieved from multiple external sources before reaching consensus.

The system should simulate oracle-like behavior such as:

- Randomized response times (to mimic network latency).

- Occasional worker failures (to simulate unreliable nodes).

- Slightly varied responses (to simulate data discrepancies across sources).

The Coordinator will use Go’s context.Context package for timeouts and partial response handling, ensuring robust behavior even if workers are delayed or fail.

Level:

Intermediate

Type of Project:

- Distributed Systems

- Go (Concurrency, Context)

- Simulation of Blockchain Oracle Behavior

Skills Required:

- Go (goroutines, context)

- Worker pool and queue design

- Error handling and resilience patterns

- API development (REST) - Gin 

Key Features:

Milestone 1: Core Worker Pool & Coordinator

- Coordinator service to accept oracle requests.

- Worker pool implementation with goroutines.

- HTTP REST communication between Coordinator and Workers.

- Result aggregation with average as default strategy.

- Timeouts via context.WithTimeout.

Milestone 2: Oracle Simulation Enhancements

- Randomized worker response times.

- Simulated worker failures (random error injection).

- Simulated variance in returned values.

- Worker metadata collection:

- Response time per worker.

- Reliability score (success vs failure rate).

Milestone 3: Fault Tolerance & Scaling

- Graceful handling when most workers fail (return partial aggregation with warning).

- Multiple concurrent oracle requests supported.

- Scale worker pool dynamically.

- Logging and reporting worker performance.

System Design

Key Components:

Coordinator

- Receives oracle requests.

- Dispatches tasks to worker pool (via channels).

- Collects and aggregates responses.

- Tracks metadata (worker performance, response times).

Workers

- Independent goroutines.

- Process tasks with randomized latency and possible failure.

- Return results with metadata.

Aggregator

- Default: average across valid worker responses.

Client Simulation

- Submits oracle requests to Coordinator.

- Displays final result with worker breakdown.

Fault Tolerance Strategy:

- If some workers fail: aggregate results from successful workers only.

- If most workers fail: return partial result with metadata flagging low reliability.

- Coordinator will record reliability scores to track worker trustworthiness over time.

Project Layout:

distributed-worker-system/
│── cmd/
│   └── main.go              # Entry point: starts Coordinator and simulates requests
│
│── pkg/
│   ├── coordinator/
│   │   ├── coordinator.go   # Coordinator logic (dispatch, collect, timeout handling)
│   │   └── aggregator.go    # Aggregation strategies (average, median, majority)
│   │
│   ├── worker/
│   │   └── worker.go        # Worker implementation (simulate latency, failures, metadata)
│   │
│   ├── models/
│   │   └── models.go        # Data models (OracleRequest, WorkerResult, OracleResult)
│   │
│   ├── client/
│   │   └── client.go        # Client simulator (submits oracle requests)
│   │
│   └── utils/
│       └── logger.go        # Helpers: logging, random generators, metadata scoring
│
│── go.mod
│── go.sum


Function Signatures (Blueprint)

📌 pkg/models/models.go

// OracleRequest represents a request to fetch data from oracles
type OracleRequest struct {
    ID    string
    Query string
}

// WorkerResult represents the response from a worker
type WorkerResult struct {
    WorkerID    int
    RequestID   string
    Value       float64
    Err         error
    ResponseTime time.Duration
    Reliable     bool
}

// OracleResult represents the final aggregated response
type OracleResult struct {
    RequestID       string
    FinalValue      float64
    WorkerResponses []WorkerResult
    ReliabilityNote string // e.g., "low reliability, most workers failed"
}

📌 pkg/worker/worker.go

// StartWorker launches a worker goroutine that listens for tasks
func StartWorker(id int, tasks <-chan models.OracleRequest, results chan<- models.WorkerResult)

// processTask simulates fetching oracle data with latency/failure
func processTask(id int, req models.OracleRequest) models.WorkerResult

// simulateResponse generates a random value with variance
func simulateResponse(base float64) float64

📌 pkg/coordinator/coordinator.go

// Coordinator manages workers, tasks, and aggregation
type Coordinator struct {
    tasks   chan models.OracleRequest
    results chan models.WorkerResult
    workers int
}

// NewCoordinator initializes coordinator with N workers
func NewCoordinator(numWorkers int) *Coordinator

// SubmitRequest submits an oracle request and waits for results (with timeout)
func (c *Coordinator) SubmitRequest(ctx context.Context, req models.OracleRequest) models.OracleResult

// dispatch sends a request to workers
func (c *Coordinator) dispatch(req models.OracleRequest)

// collectResults gathers worker responses with timeout
func (c *Coordinator) collectResults(ctx context.Context, req models.OracleRequest) []models.WorkerResult

📌 pkg/coordinator/aggregator.go
// aggregateAverage computes the average of successful worker results
func aggregateAverage(results []models.WorkerResult) float64

// aggregateMedian computes the median result
func aggregateMedian(results []models.WorkerResult) float64

// aggregateMajorityVote computes majority vote (rounding values, useful for categorical data)
func aggregateMajorityVote(results []models.WorkerResult) float64

📌 pkg/client/client.go

// SimulateClient submits a batch of requests and logs results
func SimulateClient(c *coordinator.Coordinator, queries []string)

// submitSingleRequest creates and submits one oracle request
func submitSingleRequest(c *coordinator.Coordinator, query string) models.OracleResult

📌 pkg/utils/logger.go
// LogWorkerResult logs worker responses with metadata
func LogWorkerResult(res models.WorkerResult)

// LogOracleResult logs final aggregated result
func LogOracleResult(result models.OracleResult)

// GenerateRequestID creates unique IDs for oracle requests
func GenerateRequestID() string

📌 cmd/main.go

func main() {
    // Initialize context with timeout
    // Start coordinator with N workers
    // Simulate client requests
}
