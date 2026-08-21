#!/bin/bash
set -e

REPO_DIR="/home/mahi17/Github/fintech"
cd "$REPO_DIR"

echo "=== RecoverIQ / RevenueIQ Microservice Persistent Launcher ==="

echo "Stopping any existing processes on ports 8005, 8006, 50002, 50003, 50004, 5173..."
fuser -k 8005/tcp 8006/tcp 50002/tcp 50003/tcp 50004/tcp 5173/tcp 2>/dev/null || true
sleep 1

# 1. Start failure_detector (Port 50002)
echo "Starting failure_detector on port 50002..."
cd "$REPO_DIR/failure_detector"
nohup env PORT=50002 python3 src/main.py > "$REPO_DIR/failure_detector.log" 2>&1 &
disown

# 2. Start recovery_orchestrator (Port 50003)
echo "Starting recovery_orchestrator on port 50003..."
cd "$REPO_DIR/recovery_orchestrator"
nohup env PORT=50003 python3 src/main.py > "$REPO_DIR/recovery_orchestrator.log" 2>&1 &
disown

# 3. Start reconciliation_engine (Port 50004)
echo "Starting reconciliation_engine on port 50004..."
cd "$REPO_DIR/reconciliation_engine"
nohup env PORT=50004 python3 src/main.py > "$REPO_DIR/reconciliation_engine.log" 2>&1 &
disown

# 4. Start ai_gateway (Port 8006)
echo "Starting ai_gateway on port 8006..."
cd "$REPO_DIR/ai_gateway"
nohup env HTTP_PORT=8006 python3 src/main.py > "$REPO_DIR/ai_gateway.log" 2>&1 &
disown

# 5. Start dashboard_api (Port 8005)
echo "Starting dashboard_api on port 8005..."
cd "$REPO_DIR/dashboard_api"
nohup env HTTP_PORT=8005 go run ./cmd/server/main.go > "$REPO_DIR/dashboard_api.log" 2>&1 &
disown

# 6. Start frontend (Port 5173)
echo "Starting frontend on port 5173..."
cd "$REPO_DIR/frontend"
nohup npm run dev -- --host --port 5173 > "$REPO_DIR/frontend.log" 2>&1 &
disown

sleep 3
echo "Checking service health..."
curl -s http://localhost:8005/healthz
curl -s http://localhost:8006/healthz
curl -s http://localhost:50002/healthz
curl -s http://localhost:50003/healthz
curl -s http://localhost:50004/healthz
echo ""
echo "All microservices running persistently!"
