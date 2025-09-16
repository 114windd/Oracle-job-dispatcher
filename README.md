# Distributed Worker System - Simulating Oracle Requests

A distributed worker system in Go that simulates oracle requests using HTTP-based communication. The system consists of a Coordinator that manages oracle requests and dynamically registered Workers that process tasks and return results.

## Features

- **HTTP Communication**: REST API between Coordinator and Workers
- **Dynamic Worker Registration**: Workers register themselves at runtime
- **Oracle Simulation**: Random delays, failures, and value discrepancies
- **Context-based Timeouts**: Robust handling of worker failures and delays
- **Multiple Aggregation Strategies**: Average, median, and majority vote
- **Metadata Collection**: Worker reliability, response times, and performance tracking
- **Fault Tolerance**: Graceful handling when workers fail or deregister

## Architecture

```
Worker A ---> POST /register ---> Coordinator
Worker B ---> POST /register ---> Coordinator  
Worker C ---> POST /register ---> Coordinator

Client ---> POST /request ---> Coordinator
Coordinator ---> POST /task ---> Worker A
Coordinator ---> POST /task ---> Worker B
Coordinator ---> POST /task ---> Worker C
Workers ---> POST /result (sync response) ---> Coordinator
Coordinator ---> Aggregate Results ---> Client
```

## Quick Start

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Start the Coordinator

```bash
go run cmd/coordinator/main.go
```

The coordinator will start on port 8080.

### 3. Start Workers

In separate terminals, start multiple workers:

```bash
# Worker 1
go run cmd/worker/main.go -port=8081

# Worker 2  
go run cmd/worker/main.go -port=8082

# Worker 3
go run cmd/worker/main.go -port=8083
```

### 4. Run the Demo

```bash
go run cmd/demo/main.go
```

## Manual Testing

### Register a Worker

```bash
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"id":"worker-1","endpoint":"http://localhost:8081"}'
```

### Submit an Oracle Request

```bash
curl -X POST http://localhost:8080/request \
  -H 'Content-Type: application/json' \
  -d '{"query":"BTC/USD"}'
```

### Check Worker Health

```bash
curl http://localhost:8081/health
```

## API Endpoints

### Coordinator API

- `POST /register` - Register a worker
- `POST /request` - Submit an oracle request

### Worker API

- `POST /task` - Process a task
- `GET /health` - Health check

## Project Structure

```
distributed-worker-system/
├── cmd/
│   ├── coordinator/main.go    # Coordinator entry point
│   ├── worker/main.go         # Worker entry point
│   └── demo/main.go          # Demo client
├── pkg/
│   ├── coordinator/
│   │   ├── coordinator.go     # Coordinator logic
│   │   └── aggregator.go      # Aggregation strategies
│   ├── worker/
│   │   └── worker.go          # Worker implementation
│   ├── models/
│   │   └── models.go          # Data models
│   ├── client/
│   │   └── client.go          # Client implementation
│   └── utils/
│       └── logger.go          # Utilities and logging
├── go.mod
└── README.md
```

## Configuration

### Worker Flags

- `-port`: Port for the worker server (default: 8081)
- `-coordinator`: Coordinator URL (default: http://localhost:8080)
- `-register`: Whether to register with coordinator (default: true)

### Example

```bash
go run cmd/worker/main.go -port=8081 -coordinator=http://localhost:8080 -register=true
```

## Oracle Simulation

The system simulates real oracle behavior:

- **Random Delays**: 100ms to 2s response time
- **Occasional Failures**: 10% chance of worker failure
- **Value Variance**: ±5% variation in returned values
- **Query-specific Values**: Different base values for different queries

## Aggregation Strategies

- **Average** (default): Mean of all successful responses
- **Median**: Middle value of sorted responses
- **Majority Vote**: Most common rounded value

## Fault Tolerance

- Workers that fail are excluded from aggregation
- Partial results are returned with reliability warnings
- Context-based timeouts prevent hanging requests
- Graceful degradation when most workers fail

## Development

### Building

```bash
# Build coordinator
go build -o bin/coordinator cmd/coordinator/main.go

# Build worker
go build -o bin/worker cmd/worker/main.go

# Build demo
go build -o bin/demo cmd/demo/main.go
```

### Running Tests

```bash
go test ./...
```

## License

This project is part of a learning exercise for distributed systems in Go.
