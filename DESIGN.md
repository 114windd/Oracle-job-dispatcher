📐 Design Document: Distributed Worker System – Simulating Oracle Requests (HTTP + Dynamic Registration)

1. Introduction

This document describes the technical design of a distributed worker system in Go that simulates oracle requests using HTTP-based communication.

The system consists of:

A Coordinator that manages oracle requests.

Dynamically registered Workers that process tasks and return results via HTTP.

Clients that submit oracle requests to the Coordinator.

Workers register themselves with the Coordinator at runtime, making the system extensible and fault-tolerant.

2. Goals & Non-Goals
Goals

Use HTTP for communication between Coordinator and Workers.

Allow dynamic worker registration at runtime.

Support oracle simulation (delays, failures, inconsistent results).

Use context.Context for timeouts and cancellations.

Provide aggregation strategies (default = average).

Collect metadata (worker reliability, response time).

Handle fault tolerance when workers fail or deregister.

1. System Architecture

High-Level Flow:

- Worker starts up and sends an HTTP POST /register request to the Coordinator with its endpoint info.

- Coordinator stores the worker in its registry.

- Client submits an OracleRequest via POST /request.

- Coordinator selects all available workers from its registry.

- Coordinator sends the request to each worker via HTTP POST /task.

- Workers simulate processing and respond with a WorkerResult.

- Coordinator aggregates responses and returns an OracleResult to the client.

- Workers periodically send heartbeat messages (optional future extension).

Diagram
 Worker A ---> POST /register ---> Coordinator
 Worker B ---> POST /register ---> Coordinator
 Worker C ---> POST /register ---> Coordinator

 Client ---> POST /request ---> Coordinator
 Coordinator ---> POST /task ---> Worker A
 Coordinator ---> POST /task ---> Worker B
 Coordinator ---> POST /task ---> Worker C
 Workers ---> POST /result (sync response) ---> Coordinator
 Coordinator ---> Aggregate Results ---> Client

4. Components

4.1 Coordinator

Exposes REST API endpoints:

- POST /register → register worker.

- POST /request → submit oracle request.

- Maintains a worker registry (map[WorkerID]WorkerInfo).

- Sends tasks to worker endpoints via HTTP.

- Collects results with timeout (using context.WithTimeout).

- Aggregates results and returns final response.

Key Methods:

- RegisterWorker(w WorkerInfo) error

- SubmitRequest(ctx context.Context, req OracleRequest) OracleResult

- dispatchToWorkers(ctx context.Context, req OracleRequest) []WorkerResult

4.2 Worker

Exposes an HTTP API:

- POST /task → process task and return result.

- Simulates oracle behavior:

- Random delays.

- Random failures.

- Slight value discrepancies.

- Returns WorkerResult with metadata.

Key Methods:

- startWorkerServer(port int)

- handleTask(w http.ResponseWriter, r *http.Request)

- processTask(req OracleRequest) WorkerResult

4.3 Aggregator

- Aggregates results from multiple workers.

- Default strategy = average.

- Extensible strategies: median, majority vote.

Key Methods:

- aggregateAverage(results []WorkerResult) float64

4.4 Client

- Sends oracle requests to the Coordinator (POST /request).

- Displays results with worker breakdown.

Key Methods:

- SubmitOracleRequest(query string) OracleResult

4.5 Utils

- Logging helpers.

- Worker reliability scoring.

- Unique request ID generation.

1. Data Models
type OracleRequest struct {
    ID    string `json:"id"`
    Query string `json:"query"`
}

type WorkerResult struct {
    WorkerID     string        `json:"worker_id"`
    RequestID    string        `json:"request_id"`
    Value        float64       `json:"value"`
    Err          string        `json:"err,omitempty"`
    ResponseTime time.Duration `json:"response_time"`
    Reliable     bool          `json:"reliable"`
}

type OracleResult struct {
    RequestID       string         `json:"request_id"`
    FinalValue      float64        `json:"final_value"`
    WorkerResponses []WorkerResult `json:"worker_responses"`
    ReliabilityNote string         `json:"reliability_note"`
}

type WorkerInfo struct {
    ID       string    `json:"id"`
    Endpoint string    `json:"endpoint"`
    LastSeen time.Time `json:"last_seen"`
    Reliable bool      `json:"reliable"`
}

6. Fault Tolerance

If a worker fails to respond → exclude it from aggregation.

If most workers fail → return partial result with warning in ReliabilityNote.

Coordinator periodically cleans up stale workers (LastSeen > timeout).

7. Project Layout
distributed-worker-system/
│── cmd/
│   ├── coordinator/main.go   # Starts Coordinator HTTP server
│   └── worker/main.go        # Starts Worker HTTP server
│
│── pkg/
│   ├── coordinator/
│   │   ├── coordinator.go    # Coordinator logic
│   │   └── aggregator.go     # Aggregation strategies
│   │
│   ├── worker/
│   │   └── worker.go         # Worker HTTP server
│   │
│   ├── models/
│   │   └── models.go         # Data models
│   │
│   ├── client/
│   │   └── client.go         # Client simulation
│   │
│   └── utils/
│       └── logger.go         # Logging, helpers
│
│── go.mod
│── go.sum

📡 API Specification: Distributed Worker System (HTTP + Dynamic Registration)
1. Coordinator API
1.1 Register Worker

Endpoint:
POST /register

Description:
Workers call this when they start up to register with the Coordinator.

Request (JSON):

{
  "id": "worker-1",
  "endpoint": "http://localhost:8081"
}


Response (JSON):

{
  "status": "registered",
  "message": "Worker worker-1 successfully registered"
}

1.2 Submit Oracle Request

Endpoint:

POST /request

Description:
Clients call this to request oracle data from the Coordinator. The Coordinator dispatches the request to all registered workers, aggregates responses, and returns the result.

Request (JSON):

{
  "query": "BTC/USD"
}


Response (JSON):

{
  "request_id": "req-12345",
  "final_value": 42135.50,
  "worker_responses": [
    {
      "worker_id": "worker-1",
      "request_id": "req-12345",
      "value": 42120.5,
      "response_time": "120ms",
      "reliable": true
    },
    {
      "worker_id": "worker-2",
      "request_id": "req-12345",
      "value": 42150.8,
      "response_time": "200ms",
      "reliable": true
    }
  ],
  "reliability_note": "All workers responded successfully"
}

2. Worker API
2.1 Process Task

Endpoint:
POST /task

Description:
The Coordinator calls this endpoint on each worker with the oracle request.

Request (JSON):

{
  "id": "req-12345",
  "query": "BTC/USD"
}


Response (JSON):

{
  "worker_id": "worker-1",
  "request_id": "req-12345",
  "value": 42120.5,
  "response_time": "120ms",
  "reliable": true
}

When generating Go code, follow standard Go function naming conventions:
- Exported functions (public API) must start with a capital letter (e.g., NewCoordinator, SubmitRequest).
- Unexported/internal functions must start with a lowercase letter (e.g., handleRequest, processTask).
- Use verbs for actions (e.g., RegisterWorker, StartWorkerServer).
- Use NewXxx for constructors (e.g., NewCoordinator).
- For HTTP handlers, use the convention handleXxx (e.g., handleRegister, handleTask).
- Avoid redundant names that repeat the package name (e.g., coordinator.SubmitRequest, not coordinator.CoordinatorSubmitRequest).

