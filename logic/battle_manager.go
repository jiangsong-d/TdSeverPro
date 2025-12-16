package logic

import (
	"sync"
	"towerdefense/game"
	"towerdefense/utils"
)

// BattleManager 战斗管理器
type BattleManager struct {
	battles map[string]*game.Battle
	mu      sync.RWMutex
}

var battleManager *BattleManager
var battleOnce sync.Once

// InitBattleManager 初始化战斗管理器
func InitBattleManager() *BattleManager {
	battleOnce.Do(func() {
		battleManager = &BattleManager{
			battles: make(map[string]*game.Battle),
		}
		utils.Info("战斗管理器初始化完成")
	})
	return battleManager
}

// GetBattleManager 获取战斗管理器
func GetBattleManager() *BattleManager {
	return battleManager
}

// AddBattle 添加战斗
func (bm *BattleManager) AddBattle(battle *game.Battle) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.battles[battle.ID] = battle
	utils.Info("战斗添加: %s", battle.ID)
}

// GetBattle 获取战斗
func (bm *BattleManager) GetBattle(battleID string) *game.Battle {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.battles[battleID]
}

// GetBattleByRoomID 通过房间ID获取战斗
func (bm *BattleManager) GetBattleByRoomID(roomID string) *game.Battle {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	for _, battle := range bm.battles {
		if battle.RoomID == roomID {
			return battle
		}
	}
	return nil
}

// RemoveBattle 移除战斗
func (bm *BattleManager) RemoveBattle(battleID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if battle, ok := bm.battles[battleID]; ok {
		battle.Stop()
		delete(bm.battles, battleID)
		utils.Info("战斗移除: %s", battleID)
	}
}

// GetActiveBattleCount 获取活跃战斗数
func (bm *BattleManager) GetActiveBattleCount() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return len(bm.battles)
}
