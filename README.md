# Distributed Worker System - Simulating Oracle Requests

[![CI](https://github.com/114windd/distributed-worker-system/actions/workflows/ci.yml/badge.svg)](https://github.com/114windd/distributed-worker-system/actions/workflows/ci.yml)

A distributed worker system in Go that simulates oracle requests using NATS message queue for scalable pub/sub communication. The system consists of a Coordinator that publishes oracle tasks to NATS and Workers that subscribe to process them, with HTTP API maintained for client compatibility.

## Features

- **NATS Pub/Sub Communication**: Scalable message queue between Coordinator and Workers
- **HTTP Client API**: REST API for client requests (unchanged interface)
- **Oracle Simulation**: Random delays, failures, and value discrepancies
- **Context-based Timeouts**: Robust handling of worker failures and delays
- **Multiple Aggregation Strategies**: Average, median, and majority vote
- **Metadata Collection**: Worker response times and performance tracking
- **Fault Tolerance**: Graceful handling when workers fail or disconnect
- **Horizontal Scaling**: Easy to add/remove workers without configuration changes
- **Rate Limiting**: Built-in rate limiting (10 req/sec) to prevent API abuse
- **Structured Errors**: Consistent JSON error responses with detailed information
- **Docker Support**: Complete containerization with docker-compose orchestration
- **CI/CD Pipeline**: Automated testing, linting, and building with GitHub Actions

## Architecture

```
Client ---> POST /request (HTTP) ---> Coordinator
Coordinator ---> Publish (oracle.tasks) ---> NATS
Workers ---> Subscribe (oracle.tasks) ---> Process ---> Publish (oracle.results) ---> NATS
Coordinator ---> Subscribe (oracle.results) ---> Aggregate ---> Return to Client (HTTP)
```

### NATS Subjects
- **`oracle.tasks`**: Coordinator publishes tasks, Workers subscribe
- **`oracle.results`**: Workers publish results, Coordinator subscribes

## Quick Start

### Prerequisites

- Go 1.23+ installed
- Docker (optional, for containerized setup)

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Install NATS Server

```bash
go install github.com/nats-io/nats-server/v2@latest

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin

# Verify installation
nats-server --version
```

### 3. Start NATS Server

```bash
# Start NATS server in background
nats-server --port 4222 --http_port 8222 --jetstream &

# Or start in foreground (Ctrl+C to stop)
nats-server --port 4222 --http_port 8222 --jetstream
```

### 4. Start the Coordinator

```bash
# In a new terminal
go run cmd/coordinator/main.go
```

The coordinator will start on port 8080 and connect to NATS.

### 5. Start Workers

In separate terminals, start multiple workers:

```bash
# Terminal 2: Worker 1
go run cmd/worker/main.go -port=8081

# Terminal 3: Worker 2  
go run cmd/worker/main.go -port=8082

# Terminal 4: Worker 3
go run cmd/worker/main.go -port=8083
```

### 6. Run the Demo

```bash
# In a new terminal
go run cmd/demo/main.go
```

## Alternative: Quick Test Script

Create a test script to start everything automatically:

```bash
# Create test script
cat > test_system.sh << 'EOF'
#!/bin/bash
echo "🚀 Starting Distributed Worker System Test"

# Start NATS server
echo "📡 Starting NATS server..."
nats-server --port 4222 --http_port 8222 --jetstream &
NATS_PID=$!
sleep 3

# Start coordinator
echo "🎯 Starting coordinator..."
go run cmd/coordinator/main.go &
COORD_PID=$!
sleep 3

# Start workers
echo "🔧 Starting workers..."
go run cmd/worker/main.go -port=8081 &
WORKER1_PID=$!
go run cmd/worker/main.go -port=8082 &
WORKER2_PID=$!
sleep 3

# Run demo
echo "🧪 Running demo..."
go run cmd/demo/main.go

# Cleanup
echo "🛑 Cleaning up..."
kill $COORD_PID $WORKER1_PID $WORKER2_PID $NATS_PID 2>/dev/null
EOF

# Make executable and run
chmod +x test_system.sh
./test_system.sh
```

## Docker Setup (Recommended)

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/114windd/distributed-worker-system.git
cd distributed-worker-system

# Start all services with Docker Compose
docker-compose -f docker/docker-compose.yml up --build

# Run demo client
docker-compose -f docker/docker-compose.yml run --rm demo
```

### Individual Docker Services

```bash
# Build images
docker build -f docker/Dockerfile.coordinator -t distributed-worker-coordinator .
docker build -f docker/Dockerfile.worker -t distributed-worker-worker .
docker build -f docker/Dockerfile.demo -t distributed-worker-demo .

# Start NATS server
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2.11-alpine --port 4222 --http_port 8222 --jetstream

# Start coordinator
docker run -d --name coordinator --link nats -p 8080:8080 distributed-worker-coordinator

# Start workers
docker run -d --name worker1 --link nats distributed-worker-worker ./worker -port=8081
docker run -d --name worker2 --link nats distributed-worker-worker ./worker -port=8082

# Run demo
docker run --rm --link coordinator distributed-worker-demo
```

## Manual Testing

### Submit an Oracle Request

```bash
curl -X POST http://localhost:8080/request \
  -H 'Content-Type: application/json' \
  -d '{"query":"BTC/USD"}'
```

### Check Coordinator Health

```bash
curl http://localhost:8080/health
```

### Check NATS Server Status

```bash
curl http://localhost:8222/varz
```

## API Endpoints

### Coordinator API

- `POST /request` - Submit an oracle request
- `GET /health` - Health check

### NATS Subjects

- `oracle.tasks` - Coordinator publishes tasks, Workers subscribe
- `oracle.results` - Workers publish results, Coordinator subscribes

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

- `-port`: Port for worker identification (default: 8081)

### NATS Configuration

- **Default URL**: `nats://localhost:4222`
- **HTTP Monitoring**: `http://localhost:8222`
- **JetStream**: Enabled for message persistence

### Example

```bash
go run cmd/worker/main.go -port=8081
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
- NATS provides message persistence and retry mechanisms
- Automatic reconnection to NATS server

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

## Phase 2 Features

### Security & Infrastructure
- **Rate Limiting**: Token bucket algorithm limiting requests to 10 req/sec per IP
- **Structured Errors**: Consistent JSON error responses with error codes and details
- **Docker Containerization**: Multi-stage builds for minimal, secure images
- **Health Checks**: Built-in health monitoring for all services

### CI/CD Pipeline
- **Automated Testing**: Unit tests, integration tests, and race condition detection
- **Code Quality**: Automated linting with golangci-lint and gofmt
- **Security Scanning**: Gosec security scanner for vulnerability detection
- **Docker Builds**: Automated Docker image building and caching
- **Build Artifacts**: Automatic binary generation and artifact upload

## Benefits of NATS Integration

- **Improved Scalability**: Easy horizontal scaling of workers
- **Better Decoupling**: Workers don't need to know coordinator endpoints
- **Enhanced Fault Tolerance**: NATS provides message persistence and retry
- **Simplified Architecture**: Clean pub/sub pattern replaces complex HTTP routing
- **Production Ready**: NATS is battle-tested in production environments

## Troubleshooting

### NATS Connection Issues

If you see "nats: no servers available for connection":

1. **Install NATS server**:
   ```bash
   go install github.com/nats-io/nats-server/v2@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **Start NATS server**:
   ```bash
   nats-server --port 4222 --http_port 8222 --jetstream
   ```

3. **Verify NATS is running**:
   ```bash
   curl http://localhost:8222/varz
   ```

4. **Check coordinator logs** for connection details

5. **Alternative: Use Docker**:
   ```bash
   docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2.11-alpine --port 4222 --http_port 8222 --jetstream
   ```

### Worker Not Processing Tasks

1. Ensure NATS server is running
2. Check worker logs for connection errors
3. Verify worker is subscribed to `oracle.tasks` subject

### Rate Limiting Issues

If you see "rate limit exceeded" errors:
1. Check your request frequency (limit: 10 req/sec per IP)
2. Implement client-side backoff and retry logic
3. Consider using multiple IP addresses for high-volume testing

### Docker Issues

If Docker containers fail to start:

1. **Fix Docker permissions**:
   ```bash
   # Add user to docker group
   sudo usermod -aG docker $USER
   newgrp docker
   
   # Or use sudo
   sudo docker-compose -f docker/docker-compose.yml up --build
   ```

2. **Check Docker daemon**:
   ```bash
   docker info
   sudo systemctl status docker
   ```

3. **Verify port availability**:
   ```bash
   netstat -tulpn | grep :8080
   ```

4. **Check container logs**:
   ```bash
   docker logs <container-name>
   ```

5. **Clean up Docker**:
   ```bash
   docker system prune -a
   docker-compose -f docker/docker-compose.yml down
   ```

### Common Issues

#### "nats-server: command not found"

This means NATS server is not in your PATH:

```bash
# Install NATS server
go install github.com/nats-io/nats-server/v2@latest

# Add Go bin to PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Or run directly
$(go env GOPATH)/bin/nats-server --port 4222 --http_port 8222 --jetstream
```

#### "Failed to connect to NATS: no servers available"

This means NATS server is not running:

```bash
# Check if NATS is running
curl http://localhost:8222/varz

# If not running, start it
nats-server --port 4222 --http_port 8222 --jetstream
```

#### "permission denied while trying to connect to the Docker daemon"

Fix Docker permissions:

```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or use sudo
sudo docker-compose -f docker/docker-compose.yml up --build
```

### CI/CD Pipeline Issues

If GitHub Actions fail:
1. Check the Actions tab in your repository
2. Review the specific job logs for error details
3. Ensure all dependencies are properly declared in go.mod
4. Verify Docker builds work locally before pushing

## License

This project is part of a learning exercise for distributed systems in Go with NATS message queue integration.
