@echo off
chcp 65001 >nul

echo =========================================
echo   塔防游戏服务器启动脚本
echo =========================================

REM 检查 Go 环境
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未安装 Go 环境
    echo 请访问 https://go.dev/dl/ 下载安装
    pause
    exit /b 1
)

echo Go 版本:
go version

REM 安装依赖
echo.
echo 正在安装依赖...
go mod download

REM 编译
echo.
echo 正在编译服务器...
go build -o towerdefense.exe main.go

REM 运行
echo.
echo =========================================
echo   服务器启动中...
echo =========================================
echo.
echo WebSocket 地址: ws://localhost:8080/ws
echo 健康检查: http://localhost:8080/health
echo.
echo 按 Ctrl+C 停止服务器
echo.

towerdefense.exe

pause
