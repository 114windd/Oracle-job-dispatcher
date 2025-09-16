# Quick Start Guide

## The Problem You Encountered

The error you saw (`connection refused` on port 8080) happened because:

1. **Workers try to register with coordinator on startup** - but the coordinator wasn't running
2. **Demo tries to connect to coordinator** - but the coordinator wasn't running

## Solution: Start Services in Correct Order

### Option 1: Use the Startup Script (Recommended)

```bash
# This will start everything in the correct order
./start_system.sh
```

### Option 2: Manual Step-by-Step

**Step 1: Start the Coordinator**
```bash
go run cmd/coordinator/main.go
```
Keep this running in Terminal 1.

**Step 2: Start Workers (in separate terminals)**
```bash
# Terminal 2
go run cmd/worker/main.go -port=8081

# Terminal 3  
go run cmd/worker/main.go -port=8082

# Terminal 4
go run cmd/worker/main.go -port=8083
```

**Step 3: Run the Demo**
```bash
# Terminal 5
go run cmd/demo/main.go
```

### Option 3: Use Make

```bash
# Start everything with proper ordering
make start
```

## What I Fixed

1. **Added health check endpoint** to coordinator (`GET /health`)
2. **Improved worker registration** - now fails gracefully if coordinator isn't available
3. **Enhanced demo** - now checks if coordinator is running before attempting requests
4. **Created startup script** - handles proper service startup order
5. **Reduced retry attempts** - workers now try 3 times instead of 5

## Testing Individual Components

### Test Coordinator Health
```bash
curl http://localhost:8080/health
```

### Test Worker Health
```bash
curl http://localhost:8081/health
```

### Test Worker Registration
```bash
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"id":"test-worker","endpoint":"http://localhost:8081"}'
```

### Test Oracle Request
```bash
curl -X POST http://localhost:8080/request \
  -H 'Content-Type: application/json' \
  -d '{"query":"BTC/USD"}'
```

## Troubleshooting

- **Port 8080 in use**: Kill any process using port 8080: `lsof -ti:8080 | xargs kill -9`
- **Workers can't register**: Make sure coordinator is running first
- **Demo fails**: Make sure both coordinator and workers are running

The system is now more robust and will give you clear error messages if something isn't running!
