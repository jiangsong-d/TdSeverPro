package config

import (
	"encoding/json"
	"os"
	"towerdefense/utils"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port              string  `json:"port"`
	MaxPlayers        int     `json:"max_players"`
	RoomCapacity      int     `json:"room_capacity"`
	HeartbeatInterval int     `json:"heartbeat_interval"` // 秒
	SessionTimeout    int     `json:"session_timeout"`    // 秒
	TickRate          int     `json:"tick_rate"`          // 游戏逻辑帧率
}

// GameConfig 游戏配置
type GameConfig struct {
	InitialGold       int     `json:"initial_gold"`
	InitialLife       int     `json:"initial_life"`
	WaveInterval      float64 `json:"wave_interval"`      // 波次间隔（秒）
	EnemySpawnInterval float64 `json:"enemy_spawn_interval"` // 敌人生成间隔（秒）
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string                 `json:"type"`      // txt, mysql, redis
	Settings map[string]interface{} `json:"settings"`
}

var (
	Server  ServerConfig
	Game    GameConfig
	Storage StorageConfig
)

// LoadConfig 加载配置
func LoadConfig() {
	// 默认配置
	Server = ServerConfig{
		Port:              ":8080",
		MaxPlayers:        1000,
		RoomCapacity:      4,
		HeartbeatInterval: 30,
		SessionTimeout:    120,
		TickRate:          20, // 20帧/秒
	}
	
	Game = GameConfig{
		InitialGold:        100,
		InitialLife:        20,
		WaveInterval:       5.0,
		EnemySpawnInterval: 1.0,
	}
	
	// 默认使用TXT存储
	Storage = StorageConfig{
		Type: "txt",
		Settings: map[string]interface{}{
			"data_dir": "./data",
		},
	}
	
	// 尝试从文件加载
	if file, err := os.ReadFile("config.json"); err == nil {
		var cfg map[string]interface{}
		if err := json.Unmarshal(file, &cfg); err == nil {
			utils.Info("配置文件加载成功")
		}
	} else {
		utils.Info("使用默认配置")
	}
}
