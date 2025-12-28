package network

import (
	"time"
	pb "towerdefense/proto"
	"towerdefense/utils"

	"google.golang.org/protobuf/proto"
)

// HandleProtoMessage 处理 protobuf 消息
func (s *Session) HandleProtoMessage(packet *pb.NetworkPacket) {
	msgType := pb.Cmd(packet.Cmd)
	
	switch msgType {
	case pb.Cmd_MSG_HEARTBEAT_REQ:
		s.handleProtoHeartbeat(packet.Payload)
	case pb.Cmd_MSG_LOGIN_REQ:
		s.handleProtoLogin(packet.Payload)
	case pb.Cmd_MSG_LOGOUT_REQ:
		s.handleProtoLogout(packet.Payload)
	case pb.Cmd_MSG_CREATE_ROOM_REQ:
		s.handleProtoCreateRoom(packet.Payload)
	case pb.Cmd_MSG_JOIN_ROOM_REQ:
		s.handleProtoJoinRoom(packet.Payload)
	case pb.Cmd_MSG_LEAVE_ROOM_REQ:
		s.handleProtoLeaveRoom(packet.Payload)
	case pb.Cmd_MSG_START_GAME_REQ:
		s.handleProtoStartGame(packet.Payload)
	case pb.Cmd_MSG_PLACE_TOWER_REQ:
		s.handleProtoPlaceTower(packet.Payload)
	case pb.Cmd_MSG_UPGRADE_TOWER_REQ:
		s.handleProtoUpgradeTower(packet.Payload)
	case pb.Cmd_MSG_SELL_TOWER_REQ:
		s.handleProtoSellTower(packet.Payload)
	case pb.Cmd_MSG_WAVE_START_REQ:
		s.handleProtoWaveStart(packet.Payload)
	default:
		utils.Warn("未知消息类型: %d", msgType)
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "未知的消息类型")
	}
}

// SendProtoMessage 发送 protobuf 消息
func (s *Session) SendProtoMessage(msgType pb.Cmd, message proto.Message) error {
	payload, err := proto.Marshal(message)
	if err != nil {
		utils.Error("序列化消息失败: %v", err)
		return err
	}
	
	packet := &pb.NetworkPacket{
		Cmd:       int32(msgType),
		Code:      0,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
	
	data, err := proto.Marshal(packet)
	if err != nil {
		utils.Error("序列化网络包失败: %v", err)
		return err
	}
	
	s.Send <- data
	return nil
}

// SendProtoError 发送错误消息
func (s *Session) SendProtoError(code pb.ErrorCode, message string) error {
	errResp := &pb.ErrorResponse{
		Code:    int32(code),
		Message: message,
	}
	
	return s.SendProtoMessage(pb.Cmd_MSG_ERROR, errResp)
}

// ========== 连接相关消息处理 ==========

func (s *Session) handleProtoHeartbeat(payload []byte) {
	var req pb.HeartbeatRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		utils.Error("解析心跳消息失败: %v", err)
		return
	}
	
	resp := &pb.HeartbeatResponse{
		Timestamp:  time.Now().Unix(),
		ServerTime: time.Now().Unix(),
		Ping:       int32(time.Now().Unix() - req.Timestamp),
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_HEARTBEAT_RSP, resp)
}

func (s *Session) handleProtoLogin(payload []byte) {
	var req pb.LoginRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "登录数据解析失败")
		return
	}
	
	// 验证 token
	if req.Token == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_TOKEN_INVALID, "缺少登录token")
		return
	}
	
	// 验证玩家信息
	if req.PlayerId == "" || req.PlayerName == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "玩家信息不完整")
		return
	}
	
	// 检查是否已登录
	if s.PlayerID != "" {
		s.SendProtoError(pb.ErrorCode_ERROR_ALREADY_LOGIN, "已经登录")
		return
	}
	
	// 设置玩家信息
	s.SetPlayerInfo(req.PlayerId, req.PlayerName)
	
	resp := &pb.LoginResponse{
		Success:    true,
		PlayerId:   req.PlayerId,
		PlayerName: req.PlayerName,
		PlayerInfo: &pb.PlayerBaseInfo{
			PlayerId:   req.PlayerId,
			PlayerName: req.PlayerName,
			Level:      1,
			Coin:       1000,
		},
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_LOGIN_RSP, resp)
	utils.Info("玩家 %s (%s) 登录成功", req.PlayerName, req.PlayerId)
}

func (s *Session) handleProtoLogout(payload []byte) {
	var req pb.LogoutRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		utils.Error("解析登出消息失败: %v", err)
		return
	}
	
	resp := &pb.LogoutResponse{
		Success: true,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_LOGOUT_RSP, resp)
	utils.Info("玩家 %s 登出", s.PlayerName)
	
	// 清理会话
	s.Close()
}

// ========== 房间相关消息处理 ==========

func (s *Session) handleProtoCreateRoom(payload []byte) {
	var req pb.CreateRoomRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "创建房间数据解析失败")
		return
	}
	
	// TODO: 实现房间创建逻辑
	resp := &pb.CreateRoomResponse{
		Success: true,
		RoomId:  "room_" + s.PlayerID,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_CREATE_ROOM_RSP, resp)
}

func (s *Session) handleProtoJoinRoom(payload []byte) {
	var req pb.JoinRoomRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "加入房间数据解析失败")
		return
	}
	
	// TODO: 实现加入房间逻辑
	resp := &pb.JoinRoomResponse{
		Success:  true,
		RoomInfo: &pb.RoomInfo{},
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_JOIN_ROOM_RSP, resp)
}

func (s *Session) handleProtoLeaveRoom(payload []byte) {
	var req pb.LeaveRoomRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		utils.Error("解析离开房间消息失败: %v", err)
		return
	}
	
	// TODO: 实现离开房间逻辑
	resp := &pb.LeaveRoomResponse{
		Success: true,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_LEAVE_ROOM_RSP, resp)
}

func (s *Session) handleProtoStartGame(payload []byte) {
	var req pb.StartGameRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "开始游戏数据解析失败")
		return
	}
	
	// TODO: 实现开始游戏逻辑
	resp := &pb.StartGameResponse{
		Success: true,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_START_GAME_RSP, resp)
}

// ========== 战斗相关消息处理 ==========

func (s *Session) handleProtoPlaceTower(payload []byte) {
	var req pb.PlaceTowerRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "放置防御塔数据解析失败")
		return
	}
	
	// TODO: 实现放置防御塔逻辑
	resp := &pb.PlaceTowerResponse{
		Success: true,
		TowerId: "tower_" + s.PlayerID,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_PLACE_TOWER_RSP, resp)
}

func (s *Session) handleProtoUpgradeTower(payload []byte) {
	var req pb.UpgradeTowerRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "升级防御塔数据解析失败")
		return
	}
	
	// TODO: 实现升级防御塔逻辑
	resp := &pb.UpgradeTowerResponse{
		Success: true,
		TowerId: req.TowerId,
		Level:   2,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_UPGRADE_TOWER_RSP, resp)
}

func (s *Session) handleProtoSellTower(payload []byte) {
	var req pb.SellTowerRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "出售防御塔数据解析失败")
		return
	}
	
	// TODO: 实现出售防御塔逻辑
	resp := &pb.SellTowerResponse{
		Success: true,
		TowerId: req.TowerId,
		Refund:  100,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_SELL_TOWER_RSP, resp)
}

func (s *Session) handleProtoWaveStart(payload []byte) {
	// TODO: 实现波次开始逻辑
	utils.Info("玩家 %s 请求开始波次", s.PlayerName)
}
