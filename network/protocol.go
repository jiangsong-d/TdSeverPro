package network

import (
	"encoding/json"
)

// MessageType 消息类型
type MessageType int

const (
	// 连接相关
	MsgTypeHeartbeat MessageType = 1000
	MsgTypeLogin     MessageType = 1001
	MsgTypeLogout    MessageType = 1002
	
	// 房间相关
	MsgTypeCreateRoom MessageType = 2001
	MsgTypeJoinRoom   MessageType = 2002
	MsgTypeLeaveRoom  MessageType = 2003
	MsgTypeRoomInfo   MessageType = 2004
	MsgTypeStartGame  MessageType = 2005
	
	// 战斗相关
	MsgTypePlaceTower   MessageType = 3001
	MsgTypeUpgradeTower MessageType = 3002
	MsgTypeSellTower    MessageType = 3003
	MsgTypeWaveStart    MessageType = 3004
	MsgTypeWaveComplete MessageType = 3005
	MsgTypeGameOver     MessageType = 3006
	
	// 同步相关
	MsgTypeSyncState    MessageType = 4001
	MsgTypeSyncEnemy    MessageType = 4002
	MsgTypeSyncTower    MessageType = 4003
	MsgTypeSyncDamage   MessageType = 4004
	
	// 错误消息
	MsgTypeError MessageType = 9999
)

// Message 通用消息结构
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ========== 连接相关 ==========

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	Timestamp int64 `json:"timestamp"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Timestamp int64 `json:"timestamp"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Token      string `json:"token"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success  bool   `json:"success"`
	PlayerID string `json:"player_id"`
	Message  string `json:"message"`
}

// ========== 房间相关 ==========

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	RoomName  string `json:"room_name"`
	MaxPlayer int    `json:"max_player"`
	LevelID   int    `json:"level_id"`
}

// CreateRoomResponse 创建房间响应
type CreateRoomResponse struct {
	Success bool   `json:"success"`
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	RoomID string `json:"room_id"`
}

// JoinRoomResponse 加入房间响应
type JoinRoomResponse struct {
	Success bool        `json:"success"`
	RoomID  string      `json:"room_id"`
	Players []PlayerInfo `json:"players"`
	Message string      `json:"message"`
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	IsReady    bool   `json:"is_ready"`
	IsHost     bool   `json:"is_host"`
}

// RoomInfoBroadcast 房间信息广播
type RoomInfoBroadcast struct {
	RoomID  string       `json:"room_id"`
	Players []PlayerInfo `json:"players"`
	Status  string       `json:"status"` // waiting, playing, finished
}

// StartGameRequest 开始游戏请求
type StartGameRequest struct {
	RoomID string `json:"room_id"`
}

// StartGameResponse 开始游戏响应
type StartGameResponse struct {
	Success  bool        `json:"success"`
	LevelID  int         `json:"level_id"`
	GameData GameInitData `json:"game_data"`
	Message  string      `json:"message"`
}

// GameInitData 游戏初始化数据
type GameInitData struct {
	Gold     int    `json:"gold"`
	Life     int    `json:"life"`
	MapData  string `json:"map_data"`
	WaveInfo []WaveInfo `json:"wave_info"`
}

// WaveInfo 波次信息
type WaveInfo struct {
	WaveNum     int      `json:"wave_num"`
	EnemyTypes  []int    `json:"enemy_types"`
	EnemyCounts []int    `json:"enemy_counts"`
	Reward      int      `json:"reward"`
}

// ========== 战斗相关 ==========

// PlaceTowerRequest 放置防御塔请求
type PlaceTowerRequest struct {
	TowerType int     `json:"tower_type"`
	PosX      float32 `json:"pos_x"`
	PosY      float32 `json:"pos_y"`
	PosZ      float32 `json:"pos_z"`
}

// PlaceTowerResponse 放置防御塔响应
type PlaceTowerResponse struct {
	Success  bool    `json:"success"`
	TowerID  string  `json:"tower_id"`
	Gold     int     `json:"gold"` // 剩余金币
	Message  string  `json:"message"`
}

// UpgradeTowerRequest 升级防御塔请求
type UpgradeTowerRequest struct {
	TowerID string `json:"tower_id"`
}

// UpgradeTowerResponse 升级防御塔响应
type UpgradeTowerResponse struct {
	Success bool   `json:"success"`
	TowerID string `json:"tower_id"`
	Level   int    `json:"level"`
	Gold    int    `json:"gold"`
	Message string `json:"message"`
}

// SellTowerRequest 出售防御塔请求
type SellTowerRequest struct {
	TowerID string `json:"tower_id"`
}

// SellTowerResponse 出售防御塔响应
type SellTowerResponse struct {
	Success bool   `json:"success"`
	TowerID string `json:"tower_id"`
	Gold    int    `json:"gold"`
	Message string `json:"message"`
}

// WaveStartBroadcast 波次开始广播
type WaveStartBroadcast struct {
	WaveNum int `json:"wave_num"`
}

// WaveCompleteBroadcast 波次完成广播
type WaveCompleteBroadcast struct {
	WaveNum int `json:"wave_num"`
	Reward  int `json:"reward"`
}

// GameOverBroadcast 游戏结束广播
type GameOverBroadcast struct {
	IsVictory   bool   `json:"is_victory"`
	TotalWaves  int    `json:"total_waves"`
	KillCount   int    `json:"kill_count"`
	TotalDamage int64  `json:"total_damage"`
	Score       int    `json:"score"`
}

// ========== 同步相关 ==========

// SyncStateBroadcast 状态同步广播
type SyncStateBroadcast struct {
	Gold        int           `json:"gold"`
	Life        int           `json:"life"`
	WaveNum     int           `json:"wave_num"`
	Enemies     []EnemyState  `json:"enemies"`
	Towers      []TowerState  `json:"towers"`
}

// EnemyState 敌人状态
type EnemyState struct {
	EnemyID   string  `json:"enemy_id"`
	Type      int     `json:"type"`
	HP        int     `json:"hp"`
	MaxHP     int     `json:"max_hp"`
	PosX      float32 `json:"pos_x"`
	PosY      float32 `json:"pos_y"`
	PosZ      float32 `json:"pos_z"`
	Speed     float32 `json:"speed"`
}

// TowerState 防御塔状态
type TowerState struct {
	TowerID   string  `json:"tower_id"`
	Type      int     `json:"type"`
	Level     int     `json:"level"`
	PosX      float32 `json:"pos_x"`
	PosY      float32 `json:"pos_y"`
	PosZ      float32 `json:"pos_z"`
	TargetID  string  `json:"target_id"`
}

// SyncDamageBroadcast 伤害同步广播
type SyncDamageBroadcast struct {
	TowerID  string `json:"tower_id"`
	EnemyID  string `json:"enemy_id"`
	Damage   int    `json:"damage"`
	IsCrit   bool   `json:"is_crit"`
	IsKill   bool   `json:"is_kill"`
}

// NewMessage 创建消息
func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type: msgType,
		Data: dataBytes,
	}, nil
}

// ParseMessage 解析消息
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
