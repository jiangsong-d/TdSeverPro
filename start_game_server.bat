@echo off
echo === 启动游戏区服 ===
set /p SERVER_ID="请输入区服ID (例如: 1): "
set /p SERVER_NAME="请输入区服名称 (例如: 一区): "
set /p PORT="请输入端口 (例如: 8081): "

cd /d %~dp0
go run main.go -type=game -id=%SERVER_ID% -name=%SERVER_NAME% -addr=:%PORT%
pause
