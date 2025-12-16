package network

import (
	"net/http"
	"towerdefense/utils"
	
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境需要验证Origin
		return true
	},
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Error("WebSocket升级失败: %v", err)
		return
	}
	
	// 创建新会话
	session := NewSession(conn)
	utils.Info("新连接建立: %s from %s", session.ID, r.RemoteAddr)
}
