package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"towerdefense/account"
	"towerdefense/config"
	"towerdefense/gameserver"
	"towerdefense/logic"
	"towerdefense/network"
	"towerdefense/utils"
)

var (
	serverType = flag.String("type", "account", "服务器类型: account(账号服) 或 game(游戏服)")
	serverID   = flag.Int("id", 1, "游戏服ID")
	serverName = flag.String("name", "一区", "游戏服名称")
	addr       = flag.String("addr", ":8080", "服务监听地址")
)

func main() {
	flag.Parse()
	
	// 初始化日志
	utils.InitLogger()
	
	// 根据类型启动不同服务器
	if *serverType == "account" {
		startAccountServer()
	} else if *serverType == "game" {
		startGameServer()
	} else {
		log.Fatal("未知的服务器类型，请使用 -type=account 或 -type=game")
	}
}

// startAccountServer 启动账号服务器
func startAccountServer() {
	utils.Info("=== 账号服务器启动 ===")
	
	// 初始化账号服务
	account.GetAccountServer()
	
	// 注册HTTP路由
	http.HandleFunc("/api/register", account.HandleRegister)
	http.HandleFunc("/api/login", account.HandleLogin)
	http.HandleFunc("/api/servers", account.HandleGetServerList)
	
	// 健康检查
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Account Server OK"))
	})
	
	// CORS支持
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})
	
	utils.Info("账号服务器监听地址: %s", *addr)
	utils.Info("API接口:")
	utils.Info("  - POST /api/register  注册")
	utils.Info("  - POST /api/login     登录")
	utils.Info("  - GET  /api/servers   获取区服列表")
	
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("账号服启动失败: ", err)
	}
}

// startGameServer 启动游戏服务器
func startGameServer() {
	utils.Info("=== 游戏服务器启动 ===")
	
	// 从环境变量或参数获取配置
	if envID := os.Getenv("SERVER_ID"); envID != "" {
		if id, err := strconv.Atoi(envID); err == nil {
			*serverID = id
		}
	}
	if envName := os.Getenv("SERVER_NAME"); envName != "" {
		*serverName = envName
	}
	if envAddr := os.Getenv("SERVER_ADDR"); envAddr != "" {
		*addr = envAddr
	}
	
	// 加载配置
	config.LoadConfig()
	
	// 初始化游戏服务器
	gameserver.InitGameServer(*serverID, *serverName, *addr, config.Server.MaxPlayers)
	
	// 初始化广播器（必须在管理器之前初始化）
	network.InitGameBroadcaster()
	
	// 初始化管理器
	logic.InitRoomManager()
	logic.InitBattleManager()
	
	// 注册WebSocket路由
	http.HandleFunc("/ws", network.HandleWebSocket)
	
	// 服务器信息接口
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := gameserver.GetGameServerManager().GetServerInfo()
		jsonStr := fmt.Sprintf(`{"server_id":%d,"server_name":"%s","online":%d,"max":%d}`, 
			info["server_id"], info["server_name"], info["online_players"], info["max_players"])
		w.Write([]byte(jsonStr))
	})
	
	// 健康检查接口
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Game Server OK"))
	})
	
	utils.Info("游戏服务器 [%s] 启动", *serverName)
	utils.Info("服务器ID: %d", *serverID)
	utils.Info("监听地址: %s", *addr)
	utils.Info("WebSocket: ws://localhost%s/ws", *addr)
	
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("游戏服启动失败: ", err)
	}
}
