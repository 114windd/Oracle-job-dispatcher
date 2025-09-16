#!/bin/bash

# Distributed Worker System Startup Script
echo "🚀 Starting Distributed Worker System"
echo "====================================="

# Create bin directory if it doesn't exist
mkdir -p bin

# Build all components
echo "🔨 Building components..."
go build -o bin/coordinator cmd/coordinator/main.go
go build -o bin/worker cmd/worker/main.go
go build -o bin/demo cmd/demo/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Function to cleanup background processes
cleanup() {
    echo ""
    echo "🛑 Shutting down system..."
    pkill -f "bin/coordinator" 2>/dev/null
    pkill -f "bin/worker" 2>/dev/null
    exit 0
}

# Set up signal handling
trap cleanup SIGINT SIGTERM

# Start coordinator
echo "🎯 Starting coordinator..."
./bin/coordinator &
COORDINATOR_PID=$!

# Wait for coordinator to start
echo "⏳ Waiting for coordinator to start..."
sleep 3

# Check if coordinator is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Coordinator failed to start"
    cleanup
    exit 1
fi

echo "✅ Coordinator is running!"

# Start workers
echo "🔧 Starting workers..."
./bin/worker -port=8081 &
WORKER1_PID=$!

./bin/worker -port=8082 &
WORKER2_PID=$!

./bin/worker -port=8083 &
WORKER3_PID=$!

# Wait for workers to register
echo "⏳ Waiting for workers to register..."
sleep 3

# Check worker count
WORKER_COUNT=$(curl -s http://localhost:8080/health | grep -o '"workers":[0-9]*' | cut -d':' -f2)
echo "📊 Registered workers: $WORKER_COUNT"

# Run demo
echo "🎬 Running demo..."
./bin/demo

# Keep running until interrupted
echo ""
echo "✅ System is running! Press Ctrl+C to stop."
echo "🌐 Coordinator: http://localhost:8080"
echo "🔧 Workers: http://localhost:8081, :8082, :8083"

# Wait for interrupt
wait
