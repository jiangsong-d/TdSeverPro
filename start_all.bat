@echo off
chcp 65001
echo ===================================
echo    塔防游戏服务器 - 一键启动
echo ===================================
echo.
echo 正在启动账号服务器...
start "账号服" cmd /k "go run main.go -type=account -addr=:8080"
timeout /t 2 /nobreak >nul

echo.
echo 正在启动游戏区服...
start "一区-烈焰" cmd /k "go run main.go -type=game -id=1 -name=一区-烈焰 -addr=:8081"
timeout /t 1 /nobreak >nul
start "二区-寒冰" cmd /k "go run main.go -type=game -id=2 -name=二区-寒冰 -addr=:8082"
timeout /t 1 /nobreak >nul
start "三区-雷霆" cmd /k "go run main.go -type=game -id=3 -name=三区-雷霆 -addr=:8083"

echo.
echo ===================================
echo 所有服务器已启动！
echo.
echo 账号服: http://localhost:8080
echo   - POST /api/login    登录
echo   - GET  /api/servers  获取区服列表
echo.
echo 游戏服:
echo   - 一区: ws://localhost:8081/ws
echo   - 二区: ws://localhost:8082/ws
echo   - 三区: ws://localhost:8083/ws
echo ===================================
pause
