package repository

import (
	"towerdefense/storage"
	"time"
)

// PlayerData 玩家数据模型（对应数据库表结构）
type PlayerData struct {
	PlayerID      string    `json:"player_id"`
	PlayerName    string    `json:"player_name"`
	IconID        int       `json:"icon_id"`         // 头像ID
	FrameID       int       `json:"frame_id"`        // 头像框ID
	Level         int       `json:"level"`
	Exp           int       `json:"exp"`
	Gold          int       `json:"gold"`
	Diamond       int       `json:"diamond"`
	VipLevel      int       `json:"vip_level"`
	TotalBattles  int       `json:"total_battles"`
	WinCount      int       `json:"win_count"`
	LoseCount     int       `json:"lose_count"`
	TotalKills    int       `json:"total_kills"`
	MaxWave       int       `json:"max_wave"`
	CreateTime    time.Time `json:"create_time"`
	LastLoginTime time.Time `json:"last_login_time"`
	IsNewPlayer   bool      `json:"is_new_player"`   // 是否新玩家
}

const TablePlayer = "players"

// PlayerRepository 玩家仓储
type PlayerRepository struct {
	storage storage.IStorage
}

// NewPlayerRepository 创建玩家仓储
func NewPlayerRepository() *PlayerRepository {
	return &PlayerRepository{
		storage: storage.GetStorage(),
	}
}

// Save 保存玩家数据
func (pr *PlayerRepository) Save(player *PlayerData) error {
	return pr.storage.Save(TablePlayer, player.PlayerID, player)
}

// Get 获取玩家数据
func (pr *PlayerRepository) Get(playerID string) (*PlayerData, error) {
	var player PlayerData
	err := pr.storage.Get(TablePlayer, playerID, &player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// Delete 删除玩家数据
func (pr *PlayerRepository) Delete(playerID string) error {
	return pr.storage.Delete(TablePlayer, playerID)
}

// Exists 检查玩家是否存在
func (pr *PlayerRepository) Exists(playerID string) (bool, error) {
	return pr.storage.Exists(TablePlayer, playerID)
}

// UpdateGold 更新金币
func (pr *PlayerRepository) UpdateGold(playerID string, gold int) error {
	player, err := pr.Get(playerID)
	if err != nil {
		return err
	}
	
	player.Gold = gold
	return pr.Save(player)
}

// UpdateLevel 更新等级
func (pr *PlayerRepository) UpdateLevel(playerID string, level int, exp int) error {
	player, err := pr.Get(playerID)
	if err != nil {
		return err
	}
	
	player.Level = level
	player.Exp = exp
	return pr.Save(player)
}

// AddBattleRecord 添加战斗记录
func (pr *PlayerRepository) AddBattleRecord(playerID string, isWin bool, kills int, wave int) error {
	player, err := pr.Get(playerID)
	if err != nil {
		return err
	}
	
	player.TotalBattles++
	if isWin {
		player.WinCount++
	} else {
		player.LoseCount++
	}
	player.TotalKills += kills
	
	if wave > player.MaxWave {
		player.MaxWave = wave
	}
	
	return pr.Save(player)
}

// GetTopPlayers 获取排行榜（按等级排序）
func (pr *PlayerRepository) GetTopPlayers(limit int) ([]*PlayerData, error) {
	// 获取所有玩家
	results, err := pr.storage.GetAll(TablePlayer)
	if err != nil {
		return nil, err
	}
	
	// TODO: 实现排序逻辑
	// 当使用数据库时，直接用 ORDER BY level DESC LIMIT ?
	
	players := make([]*PlayerData, 0)
	_ = results // 暂时不实现
	
	return players, nil
}

// CreateDefaultPlayer 创建默认玩家数据
func (pr *PlayerRepository) CreateDefaultPlayer(playerID, playerName string) (*PlayerData, error) {
	now := time.Now()
	player := &PlayerData{
		PlayerID:      playerID,
		PlayerName:    playerName,
		IconID:        1,        // 默认头像ID
		FrameID:       1,        // 默认头像框ID
		Level:         1,        // 默认等级
		Exp:           0,
		Gold:          1000,     // 初始金币
		Diamond:       100,      // 初始钻石
		VipLevel:      0,
		TotalBattles:  0,
		WinCount:      0,
		LoseCount:     0,
		TotalKills:    0,
		MaxWave:       0,
		CreateTime:    now,
		LastLoginTime: now,
		IsNewPlayer:   true,     // 标记为新玩家
	}
	
	err := pr.Save(player)
	if err != nil {
		return nil, err
	}
	return player, nil
}

// GetOrCreatePlayer 获取玩家数据，不存在则创建默认数据
func (pr *PlayerRepository) GetOrCreatePlayer(playerID, playerName string) (*PlayerData, bool, error) {
	player, err := pr.Get(playerID)
	if err == nil {
		// 玩家存在，更新最后登录时间
		player.LastLoginTime = time.Now()
		player.IsNewPlayer = false
		pr.Save(player)
		return player, false, nil // false 表示不是新创建的
	}
	
	// 玩家不存在，创建默认数据
	player, err = pr.CreateDefaultPlayer(playerID, playerName)
	if err != nil {
		return nil, false, err
	}
	return player, true, nil // true 表示是新创建的
}

// UpdatePlayerName 更新玩家名称
func (pr *PlayerRepository) UpdatePlayerName(playerID, newName string) error {
	player, err := pr.Get(playerID)
	if err != nil {
		return err
	}
	player.PlayerName = newName
	return pr.Save(player)
}

// UpdatePlayerIcon 更新玩家头像
func (pr *PlayerRepository) UpdatePlayerIcon(playerID string, iconID int) error {
	player, err := pr.Get(playerID)
	if err != nil {
		return err
	}
	player.IconID = iconID
	return pr.Save(player)
}
