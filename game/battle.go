package game

import (
	"sync"
	"time"
	"towerdefense/utils"
	
	"github.com/google/uuid"
)

// BattleStatus 战斗状态
type BattleStatus string

const (
	BattleStatusPreparing BattleStatus = "preparing"
	BattleStatusRunning   BattleStatus = "running"
	BattleStatusPaused    BattleStatus = "paused"
	BattleStatusFinished  BattleStatus = "finished"
)

// Battle 战斗
type Battle struct {
	ID            string
	RoomID        string
	LevelID       int
	Status        BattleStatus
	Players       map[string]*Player
	Towers        map[string]*Tower
	Enemies       map[string]*Enemy
	CurrentWave   *Wave
	WaveNum       int
	TotalWaves    int
	GameTime      float32
	LastTickTime  time.Time
	Path          []Vector3
	IsVictory     bool
	ticker        *time.Ticker
	stopChan      chan bool
	mu            sync.RWMutex
}

// NewBattle 创建战斗
func NewBattle(roomID string, levelID int, players map[string]*Player) *Battle {
	// 创建默认路径（实际应该从关卡配置加载）
	path := []Vector3{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 10},
		{X: 20, Y: 0, Z: 10},
		{X: 20, Y: 0, Z: 20},
	}
	
	return &Battle{
		ID:           uuid.New().String(),
		RoomID:       roomID,
		LevelID:      levelID,
		Status:       BattleStatusPreparing,
		Players:      players,
		Towers:       make(map[string]*Tower),
		Enemies:      make(map[string]*Enemy),
		WaveNum:      0,
		TotalWaves:   10,
		GameTime:     0,
		Path:         path,
		stopChan:     make(chan bool),
	}
}

// Start 开始战斗
func (b *Battle) Start() {
	b.mu.Lock()
	b.Status = BattleStatusRunning
	b.LastTickTime = time.Now()
	b.mu.Unlock()
	
	// 开始第一波
	b.StartNextWave()
	
	// 启动游戏循环（20帧/秒）
	b.ticker = time.NewTicker(50 * time.Millisecond)
	go b.GameLoop()
	
	utils.Info("战斗 %s 开始", b.ID)
}

// Stop 停止战斗
func (b *Battle) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.Status == BattleStatusFinished {
		return
	}
	
	b.Status = BattleStatusFinished
	if b.ticker != nil {
		b.ticker.Stop()
	}
	close(b.stopChan)
	
	utils.Info("战斗 %s 结束", b.ID)
}

// GameLoop 游戏循环
func (b *Battle) GameLoop() {
	for {
		select {
		case <-b.ticker.C:
			b.Tick()
		case <-b.stopChan:
			return
		}
	}
}

// Tick 游戏帧更新
func (b *Battle) Tick() {
	b.mu.Lock()
	if b.Status != BattleStatusRunning {
		b.mu.Unlock()
		return
	}
	
	// 计算deltaTime
	now := time.Now()
	deltaTime := float32(now.Sub(b.LastTickTime).Seconds())
	b.LastTickTime = now
	b.GameTime += deltaTime
	b.mu.Unlock()
	
	// 生成敌人
	b.SpawnEnemies()
	
	// 移动敌人
	b.MoveEnemies(deltaTime)
	
	// 塔攻击
	b.TowerAttack()
	
	// 检查波次完成
	b.CheckWaveComplete()
	
	// 定期同步状态
	if int(b.GameTime*10)%10 == 0 { // 每1秒同步一次
		b.SyncState()
	}
}

// StartNextWave 开始下一波
func (b *Battle) StartNextWave() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.WaveNum++
	if b.WaveNum > b.TotalWaves {
		// 游戏胜利
		b.IsVictory = true
		b.Status = BattleStatusFinished
		b.BroadcastGameOver()
		return
	}
	
	b.CurrentWave = NewWave(b.WaveNum)
	utils.Info("波次 %d 开始", b.WaveNum)
	
	// 广播波次开始
	b.BroadcastWaveStart()
}

// SpawnEnemies 生成敌人
func (b *Battle) SpawnEnemies() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.CurrentWave == nil {
		return
	}
	
	if b.CurrentWave.CanSpawn(b.GameTime) {
		if enemyType, ok := b.CurrentWave.GetNextEnemy(b.GameTime); ok {
			enemy := NewEnemy(uuid.New().String(), enemyType, b.Path)
			b.Enemies[enemy.ID] = enemy
		}
	}
}

// MoveEnemies 移动敌人
func (b *Battle) MoveEnemies(deltaTime float32) {
	b.mu.RLock()
	enemies := make([]*Enemy, 0, len(b.Enemies))
	for _, e := range b.Enemies {
		enemies = append(enemies, e)
	}
	b.mu.RUnlock()
	
	for _, enemy := range enemies {
		if !enemy.IsEnemyAlive() {
			continue
		}
		
		// 移动
		reachedEnd := enemy.Move(deltaTime)
		
		// 到达终点
		if reachedEnd {
			b.EnemyReachEnd(enemy)
		}
	}
}

// EnemyReachEnd 敌人到达终点
func (b *Battle) EnemyReachEnd(enemy *Enemy) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	// 减少所有玩家生命值
	for _, player := range b.Players {
		player.ReduceLife(enemy.Damage)
		
		// 检查游戏失败
		if player.GetLife() <= 0 {
			b.IsVictory = false
			b.Status = BattleStatusFinished
			b.BroadcastGameOver()
		}
	}
	
	// 移除敌人
	delete(b.Enemies, enemy.ID)
	enemy.IsAlive = false
}

// TowerAttack 塔攻击
func (b *Battle) TowerAttack() {
	b.mu.RLock()
	towers := make([]*Tower, 0, len(b.Towers))
	enemies := make([]*Enemy, 0, len(b.Enemies))
	for _, t := range b.Towers {
		towers = append(towers, t)
	}
	for _, e := range b.Enemies {
		enemies = append(enemies, e)
	}
	gameTime := b.GameTime
	b.mu.RUnlock()
	
	for _, tower := range towers {
		// 寻找目标
		tower.FindTarget(enemies)
		
		// 攻击
		if tower.CanAttack(gameTime) {
			damage, isCrit := tower.Attack(gameTime)
			
			if tower.Target != nil {
				isKill := tower.Target.TakeDamage(damage)
				
				// 广播伤害
				b.BroadcastDamage(tower.ID, tower.Target.ID, damage, isCrit, isKill)
				
				// 击杀奖励
				if isKill {
					if player := b.GetPlayer(tower.OwnerID); player != nil {
						player.AddGold(tower.Target.Gold)
						player.AddKill(1)
					}
					
					// 移除敌人
					b.mu.Lock()
					delete(b.Enemies, tower.Target.ID)
					b.mu.Unlock()
				}
			}
		}
	}
}

// CheckWaveComplete 检查波次完成
func (b *Battle) CheckWaveComplete() {
	b.mu.RLock()
	currentWave := b.CurrentWave
	enemyCount := len(b.Enemies)
	b.mu.RUnlock()
	
	if currentWave == nil || currentWave.IsWaveComplete() {
		return
	}
	
	spawned, total := currentWave.GetProgress()
	if spawned >= total && enemyCount == 0 {
		// 波次完成
		currentWave.Complete()
		
		// 发放奖励
		b.mu.Lock()
		for _, player := range b.Players {
			player.AddGold(currentWave.Reward)
		}
		b.mu.Unlock()
		
		// 广播波次完成
		b.BroadcastWaveComplete()
		
		// 延迟开始下一波
		time.AfterFunc(3*time.Second, func() {
			b.StartNextWave()
		})
	}
}

// PlaceTower 放置塔
func (b *Battle) PlaceTower(playerID string, towerType int, pos Vector3) (*Tower, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	player := b.Players[playerID]
	if player == nil {
		return nil, false
	}
	
	// 创建塔
	tower := NewTower(uuid.New().String(), towerType, playerID, pos)
	
	// 检查金币
	if !player.SpendGold(tower.Cost) {
		return nil, false
	}
	
	b.Towers[tower.ID] = tower
	return tower, true
}

// GetPlayer 获取玩家
func (b *Battle) GetPlayer(playerID string) *Player {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Players[playerID]
}

// SyncState 同步状态
func (b *Battle) SyncState() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	if globalBroadcaster == nil {
		return
	}
	
	// 收集敌人状态
	enemies := make([]EnemyState, 0, len(b.Enemies))
	for _, e := range b.Enemies {
		if !e.IsEnemyAlive() {
			continue
		}
		hp, maxHP := e.GetHP()
		pos := e.GetPosition()
		enemies = append(enemies, EnemyState{
			EnemyID: e.ID,
			Type:    e.Type,
			HP:      hp,
			MaxHP:   maxHP,
			PosX:    pos.X,
			PosY:    pos.Y,
			PosZ:    pos.Z,
			Speed:   e.Speed,
		})
	}
	
	// 收集塔状态
	towers := make([]TowerState, 0, len(b.Towers))
	for _, t := range b.Towers {
		targetID := ""
		if t.Target != nil {
			targetID = t.Target.ID
		}
		towers = append(towers, TowerState{
			TowerID:  t.ID,
			Type:     t.Type,
			Level:    t.Level,
			PosX:     t.Position.X,
			PosY:     t.Position.Y,
			PosZ:     t.Position.Z,
			TargetID: targetID,
		})
	}
	
	// 发送给每个玩家
	for _, player := range b.Players {
		sync := SyncStateBroadcast{
			Gold:    player.GetGold(),
			Life:    player.GetLife(),
			WaveNum: b.WaveNum,
			Enemies: enemies,
			Towers:  towers,
		}
		
		globalBroadcaster.BroadcastToPlayer(player.ID, MsgTypeSyncStateNtf, sync)
	}
}

// BroadcastWaveStart 广播波次开始
func (b *Battle) BroadcastWaveStart() {
	if globalBroadcaster == nil {
		return
	}
	
	broadcast := WaveStartBroadcast{
		WaveNum: b.WaveNum,
	}
	
	for _, player := range b.Players {
		globalBroadcaster.BroadcastToPlayer(player.ID, MsgTypeWaveStartRsp, broadcast)
	}
}

// BroadcastWaveComplete 广播波次完成
func (b *Battle) BroadcastWaveComplete() {
	if globalBroadcaster == nil {
		return
	}
	
	broadcast := WaveCompleteBroadcast{
		WaveNum: b.WaveNum,
		Reward:  b.CurrentWave.Reward,
	}
	
	for _, player := range b.Players {
		globalBroadcaster.BroadcastToPlayer(player.ID, MsgTypeWaveCompleteNtf, broadcast)
	}
}

// BroadcastDamage 广播伤害
func (b *Battle) BroadcastDamage(towerID, enemyID string, damage int, isCrit, isKill bool) {
	if globalBroadcaster == nil {
		return
	}
	
	broadcast := SyncDamageBroadcast{
		TowerID: towerID,
		EnemyID: enemyID,
		Damage:  damage,
		IsCrit:  isCrit,
		IsKill:  isKill,
	}
	
	for _, player := range b.Players {
		globalBroadcaster.BroadcastToPlayer(player.ID, MsgTypeSyncDamageNtf, broadcast)
	}
}

// BroadcastGameOver 广播游戏结束
func (b *Battle) BroadcastGameOver() {
	if globalBroadcaster == nil {
		return
	}
	
	totalKills := 0
	totalDamage := int64(0)
	
	for _, player := range b.Players {
		totalKills += player.KillCount
	}
	
	broadcast := GameOverBroadcast{
		IsVictory:   b.IsVictory,
		TotalWaves:  b.WaveNum,
		KillCount:   totalKills,
		TotalDamage: totalDamage,
		Score:       totalKills*10 + b.WaveNum*100,
	}
	
	for _, player := range b.Players {
		globalBroadcaster.BroadcastToPlayer(player.ID, MsgTypeGameOverNtf, broadcast)
	}
}
