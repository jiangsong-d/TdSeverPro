package repository

import (
	"towerdefense/storage"
	"time"
)

// GameRecordData 游戏记录数据模型
type GameRecordData struct {
	RecordID   string    `json:"record_id"`
	RoomID     string    `json:"room_id"`
	RoomName   string    `json:"room_name"`
	LevelID    int       `json:"level_id"`
	PlayerIDs  []string  `json:"player_ids"`
	IsWin      bool      `json:"is_win"`
	WaveCount  int       `json:"wave_count"`
	Duration   int       `json:"duration"`      // 秒
	TotalKills int       `json:"total_kills"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

const TableGameRecord = "game_records"

// GameRecordRepository 游戏记录仓储
type GameRecordRepository struct {
	storage storage.IStorage
}

// NewGameRecordRepository 创建游戏记录仓储
func NewGameRecordRepository() *GameRecordRepository {
	return &GameRecordRepository{
		storage: storage.GetStorage(),
	}
}

// Save 保存游戏记录
func (gr *GameRecordRepository) Save(record *GameRecordData) error {
	return gr.storage.Save(TableGameRecord, record.RecordID, record)
}

// Get 获取游戏记录
func (gr *GameRecordRepository) Get(recordID string) (*GameRecordData, error) {
	var record GameRecordData
	err := gr.storage.Get(TableGameRecord, recordID, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByPlayerID 获取玩家的游戏记录
func (gr *GameRecordRepository) GetByPlayerID(playerID string, limit int) ([]*GameRecordData, error) {
	// 查询所有记录
	results, err := gr.storage.GetAll(TableGameRecord)
	if err != nil {
		return nil, err
	}
	
	// TODO: 过滤和排序逻辑
	// 当使用数据库时：SELECT * FROM game_records WHERE player_ids LIKE ? ORDER BY start_time DESC LIMIT ?
	
	records := make([]*GameRecordData, 0)
	_ = results
	
	return records, nil
}

// GetRecent 获取最近的游戏记录
func (gr *GameRecordRepository) GetRecent(limit int) ([]*GameRecordData, error) {
	results, err := gr.storage.GetAll(TableGameRecord)
	if err != nil {
		return nil, err
	}
	
	// TODO: 排序逻辑
	records := make([]*GameRecordData, 0)
	_ = results
	
	return records, nil
}
