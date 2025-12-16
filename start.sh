#!/bin/bash

# 塔防服务器启动脚本

echo "========================================="
echo "  塔防游戏服务器启动脚本"
echo "========================================="

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "错误: 未安装 Go 环境"
    echo "请访问 https://go.dev/dl/ 下载安装"
    exit 1
fi

echo "Go 版本: $(go version)"

# 安装依赖
echo ""
echo "正在安装依赖..."
go mod download

# 编译
echo ""
echo "正在编译服务器..."
go build -o towerdefense main.go

# 运行
echo ""
echo "========================================="
echo "  服务器启动中..."
echo "========================================="
echo ""
echo "WebSocket 地址: ws://localhost:8080/ws"
echo "健康检查: http://localhost:8080/health"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

./towerdefense
