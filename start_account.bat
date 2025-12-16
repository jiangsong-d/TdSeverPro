@echo off
echo === 启动账号服务器 ===
cd /d %~dp0
go run main.go -type=account -addr=:8080
pause
