package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	pb "towerdefense/proto"
	"towerdefense/repository"
	"towerdefense/utils"

	"google.golang.org/protobuf/proto"
)

// 玩家仓储单例
var playerRepo = repository.NewPlayerRepository()

// 账号服地址（用于验证token）
var accountServerURL = "http://192.168.2.100:8080"

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
	case pb.Cmd_MSG_GET_PLAYER_DATA_REQ:
		s.handleProtoGetPlayerData(packet.Payload)
	case pb.Cmd_MSG_UPDATE_PLAYER_NAME_REQ:
		s.handleProtoUpdatePlayerName(packet.Payload)
	case pb.Cmd_MSG_UPDATE_PLAYER_ICON_REQ:
		s.handleProtoUpdatePlayerIcon(packet.Payload)
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

// verifyTokenViaHTTP 通过 HTTP 调用账号服验证 token
func verifyTokenViaHTTP(token string) (playerID, username string, err error) {
	reqBody, _ := json.Marshal(map[string]string{"token": token})
	resp, err := http.Post(accountServerURL+"/api/verify_token", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			PlayerID   string `json:"player_id"`
			Username   string `json:"username"`
			ExpireTime int64  `json:"expire_time"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	
	if result.Code != 0 {
		return "", "", fmt.Errorf(result.Message)
	}
	
	return result.Data.PlayerID, result.Data.Username, nil
}

// SendProtoMessage 发送 protobuf 消息
func (s *Session) SendProtoMessage(msgType pb.Cmd, message proto.Message) error {
	utils.Info("准备发送消息，Cmd: %d", msgType)
	
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
	
	utils.Info("发送消息到客户端，Cmd: %d, 数据长度: %d", msgType, len(data))
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
	utils.Info("收到登录请求，开始处理...")
	
	var req pb.LoginRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		utils.Error("登录数据解析失败: %v", err)
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "登录数据解析失败")
		return
	}
	
	utils.Info("登录请求 Token: %s, DeviceId: %s, Platform: %s", req.Token, req.DeviceId, req.Platform)
	
	// 验证 token
	if req.Token == "" {
		utils.Warn("登录失败：缺少token")
		s.SendProtoError(pb.ErrorCode_ERROR_TOKEN_INVALID, "缺少登录token")
		return
	}
	
	// 向账号服验证 token 获取玩家信息
	utils.Info("开始验证 token (通过账号服API)...")
	playerID, username, err := verifyTokenViaHTTP(req.Token)
	if err != nil {
		utils.Error("token验证失败: %v", err)
		s.SendProtoError(pb.ErrorCode_ERROR_TOKEN_INVALID, "token无效: "+err.Error())
		return
	}
	utils.Info("token验证成功，PlayerID: %s, Username: %s", playerID, username)
	
	// 检查是否已登录
	if s.PlayerID != "" {
		s.SendProtoError(pb.ErrorCode_ERROR_ALREADY_LOGIN, "已经登录")
		return
	}
	
	// 设置玩家信息（从 token 验证结果获取，同时保存 token 用于断开时清理）
	s.SetPlayerInfo(playerID, username, req.Token)
	
	resp := &pb.LoginResponse{
		Success:    true,
		PlayerId:   playerID,
		PlayerName: username,
		PlayerInfo: &pb.PlayerBaseInfo{
			PlayerId:   playerID,
			PlayerName: username,
			Level:      1,
			Coin:       1000,
		},
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_LOGIN_RSP, resp)
	utils.Info("玩家 %s (%s) 登录游戏服成功", username, playerID)
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

// ========== 玩家数据相关消息处理 ==========

func (s *Session) handleProtoGetPlayerData(payload []byte) {
	// 检查是否已登录
	if s.PlayerID == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_NOT_LOGIN, "请先登录")
		return
	}
	
	// 获取或创建玩家数据
	playerData, isNew, err := playerRepo.GetOrCreatePlayer(s.PlayerID, s.PlayerName)
	if err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_UNKNOWN, "获取玩家数据失败: "+err.Error())
		return
	}
	
	if isNew {
		utils.Info("为玩家 %s (%s) 创建了默认角色数据", s.PlayerName, s.PlayerID)
	}
	
	// 构建响应
	resp := &pb.GetPlayerDataResponse{
		Success: true,
		Message: "获取成功",
		PlayerData: &pb.PlayerData{
			PlayerId:      playerData.PlayerID,
			PlayerName:    playerData.PlayerName,
			IconId:        int32(playerData.IconID),
			FrameId:       int32(playerData.FrameID),
			Level:         int32(playerData.Level),
			Exp:           int32(playerData.Exp),
			Gold:          int32(playerData.Gold),
			Diamond:       int32(playerData.Diamond),
			VipLevel:      int32(playerData.VipLevel),
			TotalBattles:  int32(playerData.TotalBattles),
			WinCount:      int32(playerData.WinCount),
			LoseCount:     int32(playerData.LoseCount),
			TotalKills:    int32(playerData.TotalKills),
			MaxWave:       int32(playerData.MaxWave),
			CreateTime:    playerData.CreateTime.Unix(),
			LastLoginTime: playerData.LastLoginTime.Unix(),
			IsNewPlayer:   playerData.IsNewPlayer,
		},
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_GET_PLAYER_DATA_RSP, resp)
	utils.Info("玩家 %s 获取数据成功，是否新玩家: %v", s.PlayerName, isNew)
}

func (s *Session) handleProtoUpdatePlayerName(payload []byte) {
	// 检查是否已登录
	if s.PlayerID == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_NOT_LOGIN, "请先登录")
		return
	}
	
	var req pb.UpdatePlayerNameRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "请求数据解析失败")
		return
	}
	
	if req.NewName == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "名称不能为空")
		return
	}
	
	// TODO: 可以添加名称敏感词检测、长度检测等
	
	// 更新名称
	err := playerRepo.UpdatePlayerName(s.PlayerID, req.NewName)
	if err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_UNKNOWN, "更新名称失败: "+err.Error())
		return
	}
	
	// 同时更新 session 中的名称
	s.PlayerName = req.NewName
	
	resp := &pb.UpdatePlayerNameResponse{
		Success: true,
		Message: "修改成功",
		NewName: req.NewName,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_UPDATE_PLAYER_NAME_RSP, resp)
	utils.Info("玩家 %s 修改名称为: %s", s.PlayerID, req.NewName)
}

func (s *Session) handleProtoUpdatePlayerIcon(payload []byte) {
	// 检查是否已登录
	if s.PlayerID == "" {
		s.SendProtoError(pb.ErrorCode_ERROR_NOT_LOGIN, "请先登录")
		return
	}
	
	var req pb.UpdatePlayerIconRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_INVALID_PARAM, "请求数据解析失败")
		return
	}
	
	// TODO: 可以添加头像ID合法性检测（是否已解锁等）
	
	// 更新头像
	err := playerRepo.UpdatePlayerIcon(s.PlayerID, int(req.IconId))
	if err != nil {
		s.SendProtoError(pb.ErrorCode_ERROR_UNKNOWN, "更新头像失败: "+err.Error())
		return
	}
	
	resp := &pb.UpdatePlayerIconResponse{
		Success: true,
		Message: "修改成功",
		IconId:  req.IconId,
	}
	
	s.SendProtoMessage(pb.Cmd_MSG_UPDATE_PLAYER_ICON_RSP, resp)
	utils.Info("玩家 %s 修改头像为: %d", s.PlayerName, req.IconId)
}
