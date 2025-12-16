package network

import (
	"encoding/json"
	"towerdefense/config"
	"towerdefense/game"
	"towerdefense/logic"
	"towerdefense/utils"
)

// HandleMessage 处理消息
func (s *Session) HandleMessage(message []byte) {
	msg, err := ParseMessage(message)
	if err != nil {
		utils.Error("解析消息失败: %v", err)
		return
	}
	
	switch msg.Type {
	case MsgTypeHeartbeat:
		s.handleHeartbeat(msg.Data)
	case MsgTypeLogin:
		s.handleLogin(msg.Data)
	case MsgTypeCreateRoom:
		s.handleCreateRoom(msg.Data)
	case MsgTypeJoinRoom:
		s.handleJoinRoom(msg.Data)
	case MsgTypeLeaveRoom:
		s.handleLeaveRoom(msg.Data)
	case MsgTypeStartGame:
		s.handleStartGame(msg.Data)
	case MsgTypePlaceTower:
		s.handlePlaceTower(msg.Data)
	case MsgTypeUpgradeTower:
		s.handleUpgradeTower(msg.Data)
	case MsgTypeSellTower:
		s.handleSellTower(msg.Data)
	default:
		utils.Warn("未知消息类型: %d", msg.Type)
	}
}

// handleHeartbeat 处理心跳
func (s *Session) handleHeartbeat(data json.RawMessage) {
	var req HeartbeatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return
	}
	
	resp := HeartbeatResponse{
		Timestamp: req.Timestamp,
	}
	s.SendMessage(MsgTypeHeartbeat, resp)
}

// handleLogin 处理登录
func (s *Session) handleLogin(data json.RawMessage) {
	var req LoginRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("登录数据解析失败")
		return
	}
	
	// 验证token（从账号服获取的token）
	if req.Token == "" {
		s.sendError("缺少登录token")
		return
	}
	
	// TODO: 调用账号服API验证token，这里暂时简单验证
	// 实际应该: accountSrv.VerifyToken(req.Token)
	
	// 简单验证玩家信息
	if req.PlayerID == "" || req.PlayerName == "" {
		s.sendError("玩家信息不完整")
		return
	}
	
	// 设置玩家信息
	s.SetPlayerInfo(req.PlayerID, req.PlayerName)
	
	// 通知游戏服务器玩家登录
	if gameserver.GetGameServerManager() != nil {
		if !gameserver.GetGameServerManager().PlayerLogin() {
			s.sendError("服务器已满")
			return
		}
	}
	
	resp := LoginResponse{
		Success:  true,
		PlayerID: req.PlayerID,
		Message:  "登录成功",
	}
	
	s.SendMessage(MsgTypeLogin, resp)
	utils.Info("玩家 %s (%s) 登录成功", req.PlayerName, req.PlayerID)
}

// handleCreateRoom 处理创建房间
func (s *Session) handleCreateRoom(data json.RawMessage) {
	var req CreateRoomRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("创建房间数据解析失败")
		return
	}
	
	playerID, playerName := s.GetPlayerInfo()
	if playerID == "" {
		s.sendError("请先登录")
		return
	}
	
	// 创建房间
	maxPlayer := req.MaxPlayer
	if maxPlayer <= 0 || maxPlayer > config.Server.RoomCapacity {
		maxPlayer = config.Server.RoomCapacity
	}
	
	room := logic.GetRoomManager().CreateRoom(req.RoomName, maxPlayer, req.LevelID, playerID)
	
	// 创建玩家并加入房间
	player := game.NewPlayer(playerID, playerName, s.ID)
	player.IsHost = true
	room.AddPlayer(player)
	s.RoomID = room.ID
	
	resp := CreateRoomResponse{
		Success: true,
		RoomID:  room.ID,
		Message: "房间创建成功",
	}
	
	s.SendMessage(MsgTypeCreateRoom, resp)
	room.BroadcastRoomInfo()
}

// handleJoinRoom 处理加入房间
func (s *Session) handleJoinRoom(data json.RawMessage) {
	var req JoinRoomRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("加入房间数据解析失败")
		return
	}
	
	playerID, playerName := s.GetPlayerInfo()
	if playerID == "" {
		s.sendError("请先登录")
		return
	}
	
	room := logic.GetRoomManager().GetRoom(req.RoomID)
	if room == nil {
		s.sendError("房间不存在")
		return
	}
	
	// 创建玩家
	player := game.NewPlayer(playerID, playerName, s.ID)
	
	// 加入房间
	if !room.AddPlayer(player) {
		s.sendError("房间已满")
		return
	}
	
	s.RoomID = room.ID
	
	// 获取房间玩家列表
	players := room.GetPlayers()
	playerInfos := make([]PlayerInfo, len(players))
	for i, p := range players {
		playerInfos[i] = PlayerInfo{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			IsReady:    p.IsPlayerReady(),
			IsHost:     p.IsHost,
		}
	}
	
	resp := JoinRoomResponse{
		Success: true,
		RoomID:  room.ID,
		Players: playerInfos,
		Message: "加入房间成功",
	}
	
	s.SendMessage(MsgTypeJoinRoom, resp)
	room.BroadcastRoomInfo()
}

// handleLeaveRoom 处理离开房间
func (s *Session) handleLeaveRoom(data json.RawMessage) {
	playerID, _ := s.GetPlayerInfo()
	if playerID == "" {
		return
	}
	
	room := logic.GetRoomManager().GetRoomByPlayerID(playerID)
	if room == nil {
		return
	}
	
	room.RemovePlayer(playerID)
	s.RoomID = ""
	
	// 如果房间空了，移除房间
	if room.GetPlayerCount() == 0 {
		logic.GetRoomManager().RemoveRoom(room.ID)
	} else {
		room.BroadcastRoomInfo()
	}
}

// handleStartGame 处理开始游戏
func (s *Session) handleStartGame(data json.RawMessage) {
	playerID, _ := s.GetPlayerInfo()
	if playerID == "" {
		s.sendError("请先登录")
		return
	}
	
	room := logic.GetRoomManager().GetRoomByPlayerID(playerID)
	if room == nil {
		s.sendError("你不在任何房间")
		return
	}
	
	player := room.GetPlayer(playerID)
	if !player.IsHost {
		s.sendError("只有房主可以开始游戏")
		return
	}
	
	// 开始游戏
	room.StartGame(config.Game.InitialGold, config.Game.InitialLife)
	
	// 发送游戏初始化数据
	gameData := GameInitData{
		Gold:     config.Game.InitialGold,
		Life:     config.Game.InitialLife,
		MapData:  "default",
		WaveInfo: make([]WaveInfo, 10),
	}
	
	// 生成波次信息
	for i := 0; i < 10; i++ {
		gameData.WaveInfo[i] = WaveInfo{
			WaveNum:     i + 1,
			EnemyTypes:  []int{1, 2},
			EnemyCounts: []int{10, 5},
			Reward:      100 + i*50,
		}
	}
	
	resp := StartGameResponse{
		Success:  true,
		LevelID:  room.LevelID,
		GameData: gameData,
		Message:  "游戏开始",
	}
	
	// 发送给所有玩家
	room.BroadcastToRoom(MsgTypeStartGame, resp)
}

// handlePlaceTower 处理放置防御塔
func (s *Session) handlePlaceTower(data json.RawMessage) {
	var req PlaceTowerRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("放置防御塔数据解析失败")
		return
	}
	
	playerID, _ := s.GetPlayerInfo()
	if playerID == "" {
		s.sendError("请先登录")
		return
	}
	
	room := logic.GetRoomManager().GetRoomByPlayerID(playerID)
	if room == nil || room.Battle == nil {
		s.sendError("游戏未开始")
		return
	}
	
	// 放置塔
	pos := game.Vector3{X: req.PosX, Y: req.PosY, Z: req.PosZ}
	tower, success := room.Battle.PlaceTower(playerID, req.TowerType, pos)
	
	if !success {
		s.sendError("金币不足")
		return
	}
	
	player := room.Battle.GetPlayer(playerID)
	resp := PlaceTowerResponse{
		Success: true,
		TowerID: tower.ID,
		Gold:    player.GetGold(),
		Message: "防御塔放置成功",
	}
	
	s.SendMessage(MsgTypePlaceTower, resp)
}

// handleUpgradeTower 处理升级防御塔
func (s *Session) handleUpgradeTower(data json.RawMessage) {
	var req UpgradeTowerRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("升级防御塔数据解析失败")
		return
	}
	
	playerID, _ := s.GetPlayerInfo()
	room := logic.GetRoomManager().GetRoomByPlayerID(playerID)
	if room == nil || room.Battle == nil {
		s.sendError("游戏未开始")
		return
	}
	
	// 获取塔
	tower := room.Battle.Towers[req.TowerID]
	if tower == nil {
		s.sendError("防御塔不存在")
		return
	}
	
	if tower.OwnerID != playerID {
		s.sendError("这不是你的防御塔")
		return
	}
	
	player := room.Battle.GetPlayer(playerID)
	cost := tower.Cost * tower.Level / 2
	
	if !player.SpendGold(cost) {
		s.sendError("金币不足")
		return
	}
	
	tower.Upgrade()
	
	resp := UpgradeTowerResponse{
		Success: true,
		TowerID: tower.ID,
		Level:   tower.Level,
		Gold:    player.GetGold(),
		Message: "升级成功",
	}
	
	s.SendMessage(MsgTypeUpgradeTower, resp)
}

// handleSellTower 处理出售防御塔
func (s *Session) handleSellTower(data json.RawMessage) {
	var req SellTowerRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError("出售防御塔数据解析失败")
		return
	}
	
	playerID, _ := s.GetPlayerInfo()
	room := logic.GetRoomManager().GetRoomByPlayerID(playerID)
	if room == nil || room.Battle == nil {
		s.sendError("游戏未开始")
		return
	}
	
	// 获取塔
	tower := room.Battle.Towers[req.TowerID]
	if tower == nil {
		s.sendError("防御塔不存在")
		return
	}
	
	if tower.OwnerID != playerID {
		s.sendError("这不是你的防御塔")
		return
	}
	
	player := room.Battle.GetPlayer(playerID)
	player.AddGold(tower.SellValue)
	
	// 移除塔
	delete(room.Battle.Towers, req.TowerID)
	
	resp := SellTowerResponse{
		Success: true,
		TowerID: tower.ID,
		Gold:    player.GetGold(),
		Message: "出售成功",
	}
	
	s.SendMessage(MsgTypeSellTower, resp)
}

// sendError 发送错误消息
func (s *Session) sendError(message string) {
	resp := ErrorResponse{
		Code:    -1,
		Message: message,
	}
	s.SendMessage(MsgTypeError, resp)
}
