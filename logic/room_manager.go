package logic

import (
	"sync"
	"towerdefense/game"
	"towerdefense/utils"
)

// RoomManager 房间管理器
type RoomManager struct {
	rooms map[string]*game.Room
	mu    sync.RWMutex
}

var roomManager *RoomManager
var roomOnce sync.Once

// InitRoomManager 初始化房间管理器
func InitRoomManager() *RoomManager {
	roomOnce.Do(func() {
		roomManager = &RoomManager{
			rooms: make(map[string]*game.Room),
		}
		utils.Info("房间管理器初始化完成")
	})
	return roomManager
}

// GetRoomManager 获取房间管理器
func GetRoomManager() *RoomManager {
	return roomManager
}

// CreateRoom 创建房间
func (rm *RoomManager) CreateRoom(name string, maxPlayer, levelID int, hostID string) *game.Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	room := game.NewRoom(name, maxPlayer, levelID, hostID)
	rm.rooms[room.ID] = room
	
	utils.Info("房间创建: %s, 房主: %s", room.ID, hostID)
	return room
}

// GetRoom 获取房间
func (rm *RoomManager) GetRoom(roomID string) *game.Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

// RemoveRoom 移除房间
func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if room, ok := rm.rooms[roomID]; ok {
		room.FinishGame()
		delete(rm.rooms, roomID)
		utils.Info("房间移除: %s", roomID)
	}
}

// GetAllRooms 获取所有房间
func (rm *RoomManager) GetAllRooms() []*game.Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	rooms := make([]*game.Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetRoomByPlayerID 通过玩家ID获取房间
func (rm *RoomManager) GetRoomByPlayerID(playerID string) *game.Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	for _, room := range rm.rooms {
		if room.GetPlayer(playerID) != nil {
			return room
		}
	}
	return nil
}

// CleanEmptyRooms 清理空房间
func (rm *RoomManager) CleanEmptyRooms() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	for id, room := range rm.rooms {
		if room.GetPlayerCount() == 0 {
			delete(rm.rooms, id)
			utils.Info("清理空房间: %s", id)
		}
	}
}
