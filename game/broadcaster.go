package game

// MessageBroadcaster 消息广播器接口
type MessageBroadcaster interface {
	BroadcastToPlayer(playerID string, msgType int, data interface{})
}

var globalBroadcaster MessageBroadcaster

// SetMessageBroadcaster 设置消息广播器
func SetMessageBroadcaster(broadcaster MessageBroadcaster) {
	globalBroadcaster = broadcaster
}

// GetMessageBroadcaster 获取消息广播器
func GetMessageBroadcaster() MessageBroadcaster {
	return globalBroadcaster
}
