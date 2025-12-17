package network

import (
	pb "towerdefense/proto"
	"towerdefense/game"
	"google.golang.org/protobuf/proto"
)

// GameMessageBroadcaster 游戏消息广播器实现
type GameMessageBroadcaster struct{}

// BroadcastToPlayer 广播消息给玩家
func (gmb *GameMessageBroadcaster) BroadcastToPlayer(playerID string, msgType int, data interface{}) {
	session := GetSessionManager().GetSessionByPlayerID(playerID)
	if session != nil && data != nil {
		if protoMsg, ok := data.(proto.Message); ok {
			session.SendProtoMessage(pb.MessageType(msgType), protoMsg)
		}
	}
}

// InitGameBroadcaster 初始化游戏广播器
func InitGameBroadcaster() {
	broadcaster := &GameMessageBroadcaster{}
	game.SetMessageBroadcaster(broadcaster)
}
