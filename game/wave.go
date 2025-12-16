package game

import (
	"sync"
)

// Wave 波次
type Wave struct {
	WaveNum        int
	EnemyTypes     []int
	EnemyCounts    []int
	SpawnInterval  float32
	Reward         int
	TotalEnemies   int
	SpawnedEnemies int
	LastSpawnTime  float32
	IsComplete     bool
	mu             sync.RWMutex
}

// NewWave 创建波次
func NewWave(waveNum int) *Wave {
	wave := &Wave{
		WaveNum:       waveNum,
		SpawnInterval: 1.0,
		IsComplete:    false,
	}
	
	// 根据波次生成敌人配置
	switch {
	case waveNum <= 3:
		// 前3波：只有小兵
		wave.EnemyTypes = []int{1}
		wave.EnemyCounts = []int{5 + waveNum*2}
		wave.Reward = 50 + waveNum*10
	case waveNum <= 6:
		// 4-6波：小兵+中型敌人
		wave.EnemyTypes = []int{1, 2}
		wave.EnemyCounts = []int{10, waveNum - 2}
		wave.Reward = 100 + waveNum*20
	case waveNum <= 9:
		// 7-9波：更多敌人
		wave.EnemyTypes = []int{1, 2}
		wave.EnemyCounts = []int{15, waveNum}
		wave.Reward = 200 + waveNum*30
	case waveNum == 10:
		// 第10波：Boss波
		wave.EnemyTypes = []int{1, 2, 3}
		wave.EnemyCounts = []int{20, 10, 1}
		wave.Reward = 500
	default:
		// 后续波次：难度递增
		wave.EnemyTypes = []int{1, 2, 3}
		wave.EnemyCounts = []int{
			10 + waveNum*2,
			5 + waveNum,
			(waveNum - 9) / 3,
		}
		wave.Reward = 300 + waveNum*50
	}
	
	// 计算总敌人数
	for _, count := range wave.EnemyCounts {
		wave.TotalEnemies += count
	}
	
	return wave
}

// CanSpawn 是否可以生成敌人
func (w *Wave) CanSpawn(currentTime float32) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.IsComplete || w.SpawnedEnemies >= w.TotalEnemies {
		return false
	}
	
	return currentTime-w.LastSpawnTime >= w.SpawnInterval
}

// GetNextEnemy 获取下一个要生成的敌人类型
func (w *Wave) GetNextEnemy(currentTime float32) (int, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.SpawnedEnemies >= w.TotalEnemies {
		return 0, false
	}
	
	// 按照配置顺序生成敌人
	spawned := 0
	for i, count := range w.EnemyCounts {
		if spawned+count > w.SpawnedEnemies {
			w.SpawnedEnemies++
			w.LastSpawnTime = currentTime
			return w.EnemyTypes[i], true
		}
		spawned += count
	}
	
	return 0, false
}

// Complete 完成波次
func (w *Wave) Complete() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.IsComplete = true
}

// IsWaveComplete 是否完成
func (w *Wave) IsWaveComplete() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.IsComplete
}

// GetProgress 获取进度
func (w *Wave) GetProgress() (int, int) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.SpawnedEnemies, w.TotalEnemies
}
