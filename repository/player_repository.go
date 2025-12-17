package repository

import (
	"towerdefense/storage"
	"time"
)

// PlayerData 玩家数据模型（对应数据库表结构）
type PlayerData struct {
	PlayerID      string    `json:"player_id"`
	PlayerName    string    `json:"player_name"`
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
