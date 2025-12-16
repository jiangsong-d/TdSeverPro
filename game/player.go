package game

import (
	"sync"
)

// Player 玩家
type Player struct {
	ID         string
	Name       string
	SessionID  string
	IsReady    bool
	IsHost     bool
	Gold       int
	Life       int
	Score      int
	KillCount  int
	mu         sync.RWMutex
}

// NewPlayer 创建玩家
func NewPlayer(id, name, sessionID string) *Player {
	return &Player{
		ID:        id,
		Name:      name,
		SessionID: sessionID,
		IsReady:   false,
		IsHost:    false,
	}
}

// SetReady 设置准备状态
func (p *Player) SetReady(ready bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.IsReady = ready
}

// IsPlayerReady 是否准备
func (p *Player) IsPlayerReady() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.IsReady
}

// AddGold 增加金币
func (p *Player) AddGold(amount int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Gold += amount
}

// SpendGold 消耗金币
func (p *Player) SpendGold(amount int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Gold >= amount {
		p.Gold -= amount
		return true
	}
	return false
}

// GetGold 获取金币
func (p *Player) GetGold() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Gold
}

// ReduceLife 减少生命
func (p *Player) ReduceLife(amount int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Life -= amount
	if p.Life < 0 {
		p.Life = 0
	}
}

// GetLife 获取生命
func (p *Player) GetLife() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Life
}

// AddKill 增加击杀数
func (p *Player) AddKill(count int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.KillCount += count
}

// InitGameData 初始化游戏数据
func (p *Player) InitGameData(gold, life int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Gold = gold
	p.Life = life
	p.Score = 0
	p.KillCount = 0
}
