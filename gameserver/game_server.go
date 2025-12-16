package gameserver

import (
	"sync"
	"towerdefense/utils"
)

// GameServerManager 游戏服务器管理器
type GameServerManager struct {
	serverID     int
	serverName   string
	port         string
	maxPlayers   int
	onlinePlayers int
	mu           sync.RWMutex
}

var gameServerManager *GameServerManager
var gsOnce sync.Once

// InitGameServer 初始化游戏服务器
func InitGameServer(serverID int, serverName string, port string, maxPlayers int) *GameServerManager {
	gsOnce.Do(func() {
		gameServerManager = &GameServerManager{
			serverID:     serverID,
			serverName:   serverName,
			port:         port,
			maxPlayers:   maxPlayers,
			onlinePlayers: 0,
		}
		utils.Info("游戏服务器初始化: ID=%d, Name=%s, Port=%s", serverID, serverName, port)
	})
	return gameServerManager
}

// GetGameServerManager 获取游戏服务器管理器
func GetGameServerManager() *GameServerManager {
	return gameServerManager
}

// GetServerID 获取服务器ID
func (gsm *GameServerManager) GetServerID() int {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.serverID
}

// GetServerName 获取服务器名称
func (gsm *GameServerManager) GetServerName() string {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.serverName
}

// GetOnlineCount 获取在线人数
func (gsm *GameServerManager) GetOnlineCount() int {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.onlinePlayers
}

// PlayerLogin 玩家登录
func (gsm *GameServerManager) PlayerLogin() bool {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	
	if gsm.onlinePlayers >= gsm.maxPlayers {
		return false
	}
	
	gsm.onlinePlayers++
	utils.Info("玩家登录，当前在线: %d/%d", gsm.onlinePlayers, gsm.maxPlayers)
	return true
}

// PlayerLogout 玩家登出
func (gsm *GameServerManager) PlayerLogout() {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	
	if gsm.onlinePlayers > 0 {
		gsm.onlinePlayers--
	}
	utils.Info("玩家登出，当前在线: %d/%d", gsm.onlinePlayers, gsm.maxPlayers)
}

// GetServerInfo 获取服务器信息
func (gsm *GameServerManager) GetServerInfo() map[string]interface{} {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	
	return map[string]interface{}{
		"server_id":      gsm.serverID,
		"server_name":    gsm.serverName,
		"online_players": gsm.onlinePlayers,
		"max_players":    gsm.maxPlayers,
	}
}
