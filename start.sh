#!/bin/bash
set -e

echo "=== MCPFlow 一键启动 ==="
echo ""

# 检查 docker compose
if command -v docker compose &> /dev/null; then
    DC="docker compose"
elif command -v docker-compose &> /dev/null; then
    DC="docker-compose"
else
    echo "错误: 请先安装 Docker Compose"
    exit 1
fi

echo "使用: $DC"
echo ""

# 构建并启动
$DC up --build -d

echo ""
echo "=== 启动完成 ==="
echo "访问地址: http://localhost:8080"
echo ""
echo "查看日志: $DC logs -f"
echo "停止服务: $DC down"
