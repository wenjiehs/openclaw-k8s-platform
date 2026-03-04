#!/bin/bash
# OpenClaw 停止开发环境脚本

echo "=== 停止 OpenClaw 开发环境 ==="

echo "停止后端..."
pkill -f "openclaw-saas-platform.*server" 2>/dev/null && echo "  ✅ 后端已停止" || echo "  (后端未运行)"

echo "停止前端..."
pkill -f "vite.*openclaw\|npm.*dev.*openclaw" 2>/dev/null && echo "  ✅ 前端已停止" || echo "  (前端未运行)"

echo "停止 Docker 容器..."
docker stop openclaw-postgres openclaw-redis 2>/dev/null && echo "  ✅ Docker 容器已停止" || echo "  (容器未运行)"

echo ""
echo "✅ 所有服务已停止"
