package network

import (
	"sync"
	"time"
	"towerdefense/account"
	pb "towerdefense/proto"
	"towerdefense/utils"
	
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// Session 会话
type Session struct {
	ID            string
	PlayerID      string
	PlayerName    string
	Token         string    // 保存登录时使用的 token，断开时清理
	Conn          *websocket.Conn
	Send          chan []byte
	LastHeartbeat time.Time
	IsAlive       bool
	RoomID        string
	mu            sync.RWMutex
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var sessionManager *SessionManager
var once sync.Once

// GetSessionManager 获取会话管理器单例
func GetSessionManager() *SessionManager {
	once.Do(func() {
		sessionManager = &SessionManager{
			sessions: make(map[string]*Session),
		}
		// 启动心跳检测
		go sessionManager.HeartbeatChecker()
	})
	return sessionManager
}

// NewSession 创建新会话
func NewSession(conn *websocket.Conn) *Session {
	session := &Session{
		ID:            uuid.New().String(),
		Conn:          conn,
		Send:          make(chan []byte, 256),
		LastHeartbeat: time.Now(),
		IsAlive:       true,
	}
	
	GetSessionManager().AddSession(session)
	
	// 启动读写协程
	go session.ReadPump()
	go session.WritePump()
	
	return session
}

// AddSession 添加会话
func (sm *SessionManager) AddSession(session *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ID] = session
	utils.Info("新会话创建: %s", session.ID)
}

// RemoveSession 移除会话
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, ok := sm.sessions[sessionID]; ok {
		session.IsAlive = false
		close(session.Send)
		delete(sm.sessions, sessionID)
		utils.Info("会话移除: %s", sessionID)
	}
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[sessionID]
}

// GetSessionByPlayerID 通过玩家ID获取会话
func (sm *SessionManager) GetSessionByPlayerID(playerID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, session := range sm.sessions {
		if session.PlayerID == playerID {
			return session
		}
	}
	return nil
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// HeartbeatChecker 心跳检测
func (sm *SessionManager) HeartbeatChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		sm.mu.Lock()
		for id, session := range sm.sessions {
			if now.Sub(session.LastHeartbeat) > 120*time.Second {
				utils.Warn("会话超时: %s", id)
				session.Close()
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// ReadPump 读取消息
func (s *Session) ReadPump() {
	defer func() {
		s.Close()
		GetSessionManager().RemoveSession(s.ID)
	}()
	
	s.Conn.SetReadDeadline(time.Now().Add(180 * time.Second))
	s.Conn.SetPongHandler(func(string) error {
		s.Conn.SetReadDeadline(time.Now().Add(180 * time.Second))
		return nil
	})
	
	for s.IsAlive {
		_, message, err := s.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Error("WebSocket错误: %v", err)
			}
			break
		}
		
		// 更新心跳时间
		s.LastHeartbeat = time.Now()
		
		// 处理 protobuf 消息
		s.HandleProtobufMessage(message)
	}
}

// WritePump 写入消息
func (s *Session) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		s.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-s.Send:
			s.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := s.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}
			
		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close 关闭会话
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.IsAlive {
		s.IsAlive = false
		s.Conn.Close()
		
		// 通知账号服清理 token（防止内存泄漏）
		if s.Token != "" {
			account.GetAccountServer().InvalidateToken(s.Token)
			utils.Info("会话关闭，清理token: 玩家=%s", s.PlayerName)
		}
	}
}

// SetPlayerInfo 设置玩家信息
func (s *Session) SetPlayerInfo(playerID, playerName, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PlayerID = playerID
	s.PlayerName = playerName
	s.Token = token
}

// GetPlayerInfo 获取玩家信息
func (s *Session) GetPlayerInfo() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PlayerID, s.PlayerName
}

// HandleProtobufMessage 处理 protobuf 消息
func (s *Session) HandleProtobufMessage(data []byte) {
	// 解析 NetworkPacket
	packet := &pb.NetworkPacket{}
	if err := proto.Unmarshal(data, packet); err != nil {
		utils.Error("解析 NetworkPacket 失败: %v", err)
		return
	}
	
	// 委托给 proto_handler 处理
	s.HandleProtoMessage(packet)
}
