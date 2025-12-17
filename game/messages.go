package game

// 消息类型常量
const (
	MsgTypeRoomInfo     = 2004
	MsgTypeSyncState    = 4001
	MsgTypeWaveStart    = 3004
	MsgTypeWaveComplete = 3005
	MsgTypeSyncDamage   = 4004
	MsgTypeGameOver     = 3006
)

// 状态同步广播结构
type SyncStateBroadcast struct {
	Gold    int          `json:"gold"`
	Life    int          `json:"life"`
	WaveNum int          `json:"wave_num"`
	Enemies []EnemyState `json:"enemies"`
	Towers  []TowerState `json:"towers"`
}

// 敌人状态
type EnemyState struct {
	EnemyID string  `json:"enemy_id"`
	Type    int     `json:"type"`
	HP      int     `json:"hp"`
	MaxHP   int     `json:"max_hp"`
	PosX    float32 `json:"pos_x"`
	PosY    float32 `json:"pos_y"`
	PosZ    float32 `json:"pos_z"`
	Speed   float32 `json:"speed"`
}

// 防御塔状态
type TowerState struct {
	TowerID  string  `json:"tower_id"`
	Type     int     `json:"type"`
	Level    int     `json:"level"`
	PosX     float32 `json:"pos_x"`
	PosY     float32 `json:"pos_y"`
	PosZ     float32 `json:"pos_z"`
	TargetID string  `json:"target_id"`
}

// 波次开始广播
type WaveStartBroadcast struct {
	WaveNum int `json:"wave_num"`
}

// 波次完成广播
type WaveCompleteBroadcast struct {
	WaveNum int `json:"wave_num"`
	Reward  int `json:"reward"`
}

// 伤害同步广播
type SyncDamageBroadcast struct {
	TowerID string `json:"tower_id"`
	EnemyID string `json:"enemy_id"`
	Damage  int    `json:"damage"`
	IsCrit  bool   `json:"is_crit"`
	IsKill  bool   `json:"is_kill"`
}

// 游戏结束广播
type GameOverBroadcast struct {
	IsVictory   bool   `json:"is_victory"`
	TotalWaves  int    `json:"total_waves"`
	KillCount   int    `json:"kill_count"`
	TotalDamage int64  `json:"total_damage"`
	Score       int    `json:"score"`
}

// 玩家信息
type PlayerInfo struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	IsReady    bool   `json:"is_ready"`
	IsHost     bool   `json:"is_host"`
}

// 房间信息广播
type RoomInfoBroadcast struct {
	RoomID  string       `json:"room_id"`
	Players []PlayerInfo `json:"players"`
	Status  string       `json:"status"` // waiting, playing, finished
}
