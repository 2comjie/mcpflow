#!/bin/bash
set -e

echo "=== MCPFlow 启动 ==="

# 构建并启动所有服务
docker compose up -d --build

echo ""
echo "服务已启动:"
echo "  前端:  http://localhost"
echo "  后端:  http://localhost:80/api/v1"
echo "  MongoDB: localhost:27017 (仅内部访问)"
echo ""
echo "查看日志: docker compose logs -f"
echo "停止服务: docker compose down"
