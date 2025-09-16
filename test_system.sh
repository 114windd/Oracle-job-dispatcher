#!/bin/bash

# Test script for Distributed Worker System
echo "🧪 Testing Distributed Worker System"
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

# Test coordinator startup
echo "🚀 Testing coordinator startup..."
timeout 5s ./bin/coordinator &
COORDINATOR_PID=$!

# Wait for coordinator to start
sleep 2

# Test worker registration
echo "📝 Testing worker registration..."
curl -s -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"id":"test-worker-1","endpoint":"http://localhost:8081"}' > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Worker registration successful"
else
    echo "❌ Worker registration failed"
fi

# Test oracle request
echo "🎯 Testing oracle request..."
RESPONSE=$(curl -s -X POST http://localhost:8080/request \
  -H 'Content-Type: application/json' \
  -d '{"query":"BTC/USD"}')

if [ $? -eq 0 ] && echo "$RESPONSE" | grep -q "final_value"; then
    echo "✅ Oracle request successful"
    echo "📊 Response: $RESPONSE"
else
    echo "❌ Oracle request failed"
fi

# Cleanup
echo "🧹 Cleaning up..."
kill $COORDINATOR_PID 2>/dev/null

echo "🎉 Test completed!"
