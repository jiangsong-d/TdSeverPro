package game

import (
	"sync"
	"towerdefense/utils"
	
	"github.com/google/uuid"
)

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

// Room 房间
type Room struct {
	ID        string
	Name      string
	MaxPlayer int
	LevelID   int
	Status    RoomStatus
	Players   map[string]*Player
	HostID    string
	Battle    *Battle
	mu        sync.RWMutex
}

// NewRoom 创建房间
func NewRoom(name string, maxPlayer, levelID int, hostID string) *Room {
	return &Room{
		ID:        uuid.New().String(),
		Name:      name,
		MaxPlayer: maxPlayer,
		LevelID:   levelID,
		Status:    RoomStatusWaiting,
		Players:   make(map[string]*Player),
		HostID:    hostID,
	}
}

// AddPlayer 添加玩家
func (r *Room) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if len(r.Players) >= r.MaxPlayer {
		return false
	}
	
	// 第一个玩家是房主
	if len(r.Players) == 0 {
		player.IsHost = true
		r.HostID = player.ID
	}
	
	r.Players[player.ID] = player
	utils.Info("玩家 %s 加入房间 %s", player.Name, r.ID)
	return true
}

// RemovePlayer 移除玩家
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.Players[playerID]; ok {
		delete(r.Players, playerID)
		utils.Info("玩家 %s 离开房间 %s", playerID, r.ID)
		
		// 如果是房主离开，转移房主
		if r.HostID == playerID && len(r.Players) > 0 {
			for _, p := range r.Players {
				p.IsHost = true
				r.HostID = p.ID
				break
			}
		}
	}
}

// GetPlayer 获取玩家
func (r *Room) GetPlayer(playerID string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Players[playerID]
}

// GetPlayers 获取所有玩家
func (r *Room) GetPlayers() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	players := make([]*Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, p)
	}
	return players
}

// GetPlayerCount 获取玩家数量
func (r *Room) GetPlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Players)
}

// AllReady 是否所有玩家准备
func (r *Room) AllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if len(r.Players) == 0 {
		return false
	}
	
	for _, p := range r.Players {
		if !p.IsPlayerReady() && !p.IsHost {
			return false
		}
	}
	return true
}

// StartGame 开始游戏
func (r *Room) StartGame(initialGold, initialLife int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.Status = RoomStatusPlaying
	
	// 初始化玩家游戏数据
	for _, p := range r.Players {
		p.InitGameData(initialGold, initialLife)
	}
	
	// 创建战斗实例
	r.Battle = NewBattle(r.ID, r.LevelID, r.Players)
	r.Battle.Start()
	
	utils.Info("房间 %s 游戏开始", r.ID)
}

// FinishGame 结束游戏
func (r *Room) FinishGame() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.Status = RoomStatusFinished
	if r.Battle != nil {
		r.Battle.Stop()
	}
	
	utils.Info("房间 %s 游戏结束", r.ID)
}

// BroadcastRoomInfo 广播房间信息
func (r *Room) BroadcastRoomInfo() {
	if globalBroadcaster == nil {
		return
	}
	
	players := r.GetPlayers()
	playerInfos := make([]PlayerInfo, len(players))
	
	for i, p := range players {
		playerInfos[i] = PlayerInfo{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			IsReady:    p.IsPlayerReady(),
			IsHost:     p.IsHost,
		}
	}
	
	broadcast := RoomInfoBroadcast{
		RoomID:  r.ID,
		Players: playerInfos,
		Status:  string(r.Status),
	}
	
	// 发送给房间内所有玩家
	for _, p := range players {
		globalBroadcaster.BroadcastToPlayer(p.ID, MsgTypeRoomInfoRsp, broadcast)
	}
}

// BroadcastToRoom 向房间广播消息
func (r *Room) BroadcastToRoom(msgType int, data interface{}) {
	if globalBroadcaster == nil {
		return
	}
	
	players := r.GetPlayers()
	for _, p := range players {
		globalBroadcaster.BroadcastToPlayer(p.ID, msgType, data)
	}
}
