package game

import (
	"math"
	"sync"
)

// Enemy 敌人
type Enemy struct {
	ID            string
	Type          int
	HP            int
	MaxHP         int
	Speed         float32
	Position      Vector3
	PathIndex     int
	Path          []Vector3
	IsAlive       bool
	Gold          int  // 击杀奖励
	Damage        int  // 到达终点的伤害
	mu            sync.RWMutex
}

// NewEnemy 创建敌人
func NewEnemy(id string, enemyType int, path []Vector3) *Enemy {
	enemy := &Enemy{
		ID:        id,
		Type:      enemyType,
		Path:      path,
		PathIndex: 0,
		IsAlive:   true,
	}
	
	if len(path) > 0 {
		enemy.Position = path[0]
	}
	
	// 根据类型设置属性
	switch enemyType {
	case 1: // 小兵
		enemy.MaxHP = 50
		enemy.HP = 50
		enemy.Speed = 2.0
		enemy.Gold = 10
		enemy.Damage = 1
	case 2: // 中型敌人
		enemy.MaxHP = 100
		enemy.HP = 100
		enemy.Speed = 1.5
		enemy.Gold = 20
		enemy.Damage = 2
	case 3: // Boss
		enemy.MaxHP = 500
		enemy.HP = 500
		enemy.Speed = 1.0
		enemy.Gold = 100
		enemy.Damage = 5
	default:
		enemy.MaxHP = 50
		enemy.HP = 50
		enemy.Speed = 2.0
		enemy.Gold = 10
		enemy.Damage = 1
	}
	
	return enemy
}

// TakeDamage 受到伤害
func (e *Enemy) TakeDamage(damage int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.HP -= damage
	if e.HP <= 0 {
		e.HP = 0
		e.IsAlive = false
		return true // 死亡
	}
	return false
}

// Move 移动
func (e *Enemy) Move(deltaTime float32) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.IsAlive || e.PathIndex >= len(e.Path) {
		return false
	}
	
	// 目标位置
	target := e.Path[e.PathIndex]
	
	// 计算方向
	dx := target.X - e.Position.X
	dy := target.Y - e.Position.Y
	dz := target.Z - e.Position.Z
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	
	// 到达当前路径点
	if dist < 0.1 {
		e.PathIndex++
		if e.PathIndex >= len(e.Path) {
			return true // 到达终点
		}
		return false
	}
	
	// 移动
	moveDistance := e.Speed * deltaTime
	if moveDistance > dist {
		moveDistance = dist
	}
	
	e.Position.X += (dx / dist) * moveDistance
	e.Position.Y += (dy / dist) * moveDistance
	e.Position.Z += (dz / dist) * moveDistance
	
	return false
}

// GetPosition 获取位置
func (e *Enemy) GetPosition() Vector3 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Position
}

// GetHP 获取血量
func (e *Enemy) GetHP() (int, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.HP, e.MaxHP
}

// IsEnemyAlive 是否存活
func (e *Enemy) IsEnemyAlive() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.IsAlive
}
