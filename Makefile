# Distributed Worker System Makefile

.PHONY: build clean test run-coordinator run-worker run-demo help

# Default target
all: build

# Build all components
build:
	@echo "🔨 Building all components..."
	@mkdir -p bin
	go build -o bin/coordinator cmd/coordinator/main.go
	go build -o bin/worker cmd/worker/main.go
	go build -o bin/demo cmd/demo/main.go
	@echo "✅ Build complete"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Run coordinator
run-coordinator: build
	@echo "🚀 Starting coordinator..."
	./bin/coordinator

# Run worker (with port argument)
run-worker: build
	@echo "🔧 Starting worker on port $(PORT)..."
	@if [ -z "$(PORT)" ]; then echo "❌ Please specify PORT=8081"; exit 1; fi
	./bin/worker -port=$(PORT)

# Run demo
run-demo: build
	@echo "🎬 Running demo..."
	./bin/demo

# Start full system (coordinator + 3 workers)
start-system: build
	@echo "🚀 Starting full system..."
	@echo "Starting coordinator..."
	./bin/coordinator &
	@sleep 2
	@echo "Starting workers..."
	./bin/worker -port=8081 &
	./bin/worker -port=8082 &
	./bin/worker -port=8083 &
	@sleep 2
	@echo "Running demo..."
	./bin/demo
	@echo "🛑 Stopping system..."
	@pkill -f coordinator
	@pkill -f worker

# Start system with proper startup script
start: build
	@echo "🚀 Starting system with startup script..."
	./start_system.sh

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Show help
help:
	@echo "Distributed Worker System - Available commands:"
	@echo ""
	@echo "  build          - Build all components"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  run-coordinator - Run coordinator server"
	@echo "  run-worker     - Run worker (requires PORT=8081)"
	@echo "  run-demo       - Run demo client"
	@echo "  start-system   - Start full system (coordinator + 3 workers + demo)"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run-coordinator"
	@echo "  make run-worker PORT=8081"
	@echo "  make start-system"
