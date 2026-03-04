#!/bin/bash
# OpenClaw 本地开发环境启动脚本

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
BACKEND_LOG="$PROJECT_ROOT/tmp/backend.log"
FRONTEND_LOG="$PROJECT_ROOT/tmp/frontend.log"

mkdir -p "$PROJECT_ROOT/tmp"

echo "=== OpenClaw 本地开发环境 ==="
echo ""

# 1. 检查/启动 Docker 容器
echo "[1/4] 检查 Docker 容器..."

if ! docker ps --format "{{.Names}}" | grep -q "openclaw-postgres"; then
  echo "  启动 PostgreSQL..."
  docker run -d \
    --name openclaw-postgres \
    -e POSTGRES_DB=openclaw \
    -e POSTGRES_USER=openclaw \
    -e POSTGRES_PASSWORD=openclaw_dev_password \
    -v openclaw-postgres-data:/var/lib/postgresql/data \
    -v "$PROJECT_ROOT/backend/migrations:/docker-entrypoint-initdb.d" \
    -p 5433:5432 \
    postgres:15-alpine
  echo "  等待 PostgreSQL 就绪..."
  sleep 5
else
  echo "  ✅ PostgreSQL 已运行 (port 5433)"
fi

if ! docker ps --format "{{.Names}}" | grep -q "openclaw-redis"; then
  echo "  启动 Redis..."
  docker run -d \
    --name openclaw-redis \
    -p 6379:6379 \
    redis:7-alpine
  sleep 2
else
  echo "  ✅ Redis 已运行 (port 6379)"
fi

# 2. 启动后端
echo ""
echo "[2/4] 启动后端服务..."
pkill -f "openclaw-saas-platform.*server" 2>/dev/null || true
sleep 1

cd "$PROJECT_ROOT/backend"
DATABASE_URL="postgres://openclaw:openclaw_dev_password@localhost:5433/openclaw?sslmode=disable" \
REDIS_URL="redis://localhost:6379" \
JWT_SECRET="dev-jwt-secret-change-in-production" \
SERVER_PORT=8080 \
SERVER_MODE=debug \
nohup go run ./cmd/server > "$BACKEND_LOG" 2>&1 &

echo "  等待后端启动..."
for i in $(seq 1 15); do
  sleep 1
  if curl -s --max-time 1 http://localhost:8080/health > /dev/null 2>&1; then
    echo "  ✅ 后端已启动 (port 8080)"
    break
  fi
  if [ $i -eq 15 ]; then
    echo "  ❌ 后端启动超时，查看日志: $BACKEND_LOG"
    tail -20 "$BACKEND_LOG"
    exit 1
  fi
done

# 3. 启动前端
echo ""
echo "[3/4] 启动前端服务..."
pkill -f "vite.*openclaw" 2>/dev/null || true
sleep 1

cd "$PROJECT_ROOT/frontend"
VITE_API_BASE_URL=http://localhost:8080 \
nohup npm run dev > "$FRONTEND_LOG" 2>&1 &

echo "  等待前端启动..."
for i in $(seq 1 20); do
  sleep 1
  if curl -s --max-time 1 http://localhost:3000 > /dev/null 2>&1; then
    echo "  ✅ 前端已启动 (port 3000)"
    break
  fi
  if [ $i -eq 20 ]; then
    echo "  ❌ 前端启动超时，查看日志: $FRONTEND_LOG"
    tail -20 "$FRONTEND_LOG"
    exit 1
  fi
done

# 4. 完成
echo ""
echo "[4/4] 所有服务已就绪 🎉"
echo ""
echo "  前端:      http://localhost:3000"
echo "  后端 API:  http://localhost:8080"
echo "  健康检查:  http://localhost:8080/health"
echo ""
echo "  后端日志:  tail -f $BACKEND_LOG"
echo "  前端日志:  tail -f $FRONTEND_LOG"
echo ""
echo "  停止服务:  ./stop-dev.sh"
