package game

import (
	"math"
	"sync"
)

// Vector3 三维向量
type Vector3 struct {
	X float32
	Y float32
	Z float32
}

// Distance 计算距离
func (v Vector3) Distance(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Tower 防御塔
type Tower struct {
	ID         string
	Type       int
	Level      int
	OwnerID    string
	Position   Vector3
	Damage     int
	AttackSpeed float32
	Range      float32
	Target     *Enemy
	LastAttack float32
	Cost       int
	SellValue  int
	mu         sync.RWMutex
}

// NewTower 创建防御塔
func NewTower(id string, towerType int, ownerID string, pos Vector3) *Tower {
	// 根据类型设置属性（这里简化处理，实际应从配置表读取）
	tower := &Tower{
		ID:       id,
		Type:     towerType,
		Level:    1,
		OwnerID:  ownerID,
		Position: pos,
	}
	
	// 不同类型塔的属性
	switch towerType {
	case 1: // 箭塔
		tower.Damage = 10
		tower.AttackSpeed = 1.0
		tower.Range = 5.0
		tower.Cost = 50
		tower.SellValue = 25
	case 2: // 炮塔
		tower.Damage = 30
		tower.AttackSpeed = 2.0
		tower.Range = 6.0
		tower.Cost = 100
		tower.SellValue = 50
	case 3: // 魔法塔
		tower.Damage = 15
		tower.AttackSpeed = 0.8
		tower.Range = 7.0
		tower.Cost = 80
		tower.SellValue = 40
	default:
		tower.Damage = 10
		tower.AttackSpeed = 1.0
		tower.Range = 5.0
		tower.Cost = 50
		tower.SellValue = 25
	}
	
	return tower
}

// FindTarget 寻找目标
func (t *Tower) FindTarget(enemies []*Enemy) *Enemy {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	var closest *Enemy
	minDist := float32(math.MaxFloat32)
	
	for _, enemy := range enemies {
		if !enemy.IsAlive {
			continue
		}
		
		dist := t.Position.Distance(enemy.Position)
		if dist <= t.Range && dist < minDist {
			closest = enemy
			minDist = dist
		}
	}
	
	t.Target = closest
	return closest
}

// CanAttack 是否可以攻击
func (t *Tower) CanAttack(currentTime float32) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.Target == nil || !t.Target.IsAlive {
		return false
	}
	
	if currentTime-t.LastAttack < t.AttackSpeed {
		return false
	}
	
	dist := t.Position.Distance(t.Target.Position)
	return dist <= t.Range
}

// Attack 攻击
func (t *Tower) Attack(currentTime float32) (int, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.LastAttack = currentTime
	
	// 简单的暴击计算
	isCrit := false
	damage := t.Damage
	if math.Mod(float64(currentTime*100), 10) < 1 { // 10%暴击率
		isCrit = true
		damage = int(float32(damage) * 2.0)
	}
	
	return damage, isCrit
}

// Upgrade 升级
func (t *Tower) Upgrade() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.Level++
	
	// 升级属性提升
	t.Damage = int(float32(t.Damage) * 1.3)
	t.AttackSpeed *= 0.9
	t.Range += 0.5
	
	cost := t.Cost * t.Level / 2
	t.SellValue = cost / 2
	
	return cost
}

// GetInfo 获取信息
func (t *Tower) GetInfo() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return map[string]interface{}{
		"id":           t.ID,
		"type":         t.Type,
		"level":        t.Level,
		"damage":       t.Damage,
		"attack_speed": t.AttackSpeed,
		"range":        t.Range,
		"cost":         t.Cost,
		"sell_value":   t.SellValue,
	}
}
