package account

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"towerdefense/repository"
	"towerdefense/utils"
	
	"github.com/google/uuid"
)

// AccountServer 账号服务器
type AccountServer struct {
	accounts       map[string]*Account  // username -> account (内存缓存)
	tokens         map[string]*Session  // token -> session
	gameServers    []*GameServerInfo    // 区服列表
	accountRepo    *repository.AccountRepository  // 账号仓储
	mu             sync.RWMutex
}

// Account 账号信息
type Account struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`      // MD5加密
	PlayerID     string    `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	CreateTime   time.Time `json:"create_time"`
	LastLoginTime time.Time `json:"last_login_time"`
}

// Session 会话信息
type Session struct {
	Token      string    `json:"token"`
	PlayerID   string    `json:"player_id"`
	Username   string    `json:"username"`
	ExpireTime time.Time `json:"expire_time"`
}

// GameServerInfo 游戏区服信息
type GameServerInfo struct {
	ServerID   int    `json:"server_id"`
	ServerName string `json:"server_name"`
	Url        string `json:"url"`         // 完整的WebSocket连接地址
	Status     string `json:"status"`      // online, maintain, full
	OnlineNum  int    `json:"online_num"`
	MaxPlayer  int    `json:"max_player"`
	Recommend  bool   `json:"recommend"`   // 推荐服
	IsNew      bool   `json:"is_new"`      // 新服
}

var accountServer *AccountServer
var accountOnce sync.Once

// GetAccountServer 获取账号服务器单例
func GetAccountServer() *AccountServer {
	accountOnce.Do(func() {
		accountServer = &AccountServer{
			accounts:    make(map[string]*Account),
			tokens:      make(map[string]*Session),
			accountRepo: repository.NewAccountRepository(),
		}
		accountServer.InitGameServers()
		accountServer.LoadAccountsFromStorage()
		
		// 启动定时清理过期 token 的协程
		go accountServer.tokenCleanupRoutine()
		
		utils.Info("账号服务器初始化完成")
	})
	return accountServer
}

// tokenCleanupRoutine 定时清理过期 token（防止内存泄漏）
func (as *AccountServer) tokenCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次
	defer ticker.Stop()
	
	for range ticker.C {
		as.cleanupExpiredTokens()
	}
}

// cleanupExpiredTokens 清理过期的 token
func (as *AccountServer) cleanupExpiredTokens() {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	now := time.Now()
	expiredCount := 0
	
	for token, session := range as.tokens {
		if now.After(session.ExpireTime) {
			delete(as.tokens, token)
			expiredCount++
		}
	}
	
	if expiredCount > 0 {
		utils.Info("清理过期token: %d 个，剩余: %d 个", expiredCount, len(as.tokens))
	}
}

// LoadAccountsFromStorage 从存储加载账号数据到内存缓存
func (as *AccountServer) LoadAccountsFromStorage() {
	// 启动时加载所有账号到内存（可选优化）
	// 对于大量账号，可以按需加载
	utils.Info("账号数据已就绪，使用持久化存储")
}

// InitGameServers 初始化区服列表（实际应该从数据库或配置文件加载）
func (as *AccountServer) InitGameServers() {
	as.gameServers = []*GameServerInfo{
		{
			ServerID:   1,
			ServerName: "一区-烈焰",
			Url:        "ws://192.168.2.100:8081/game/login",
			Status:     "online",
			OnlineNum:  125,
			MaxPlayer:  1000,
			Recommend:  true,
			IsNew:      false,
		},
	}
}

// RegisterAccount 注册账号
func (as *AccountServer) RegisterAccount(username, password string) (*Account, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	// 检查用户名是否存在（从存储检查）
	exists, err := as.accountRepo.Exists(username)
	if err == nil && exists {
		return nil, fmt.Errorf("用户名已存在")
	}
	
	// 创建账号
	account := &Account{
		Username:      username,
		Password:      hashPassword(password),
		PlayerID:      uuid.New().String(),
		PlayerName:    username,
		CreateTime:    time.Now(),
		LastLoginTime: time.Now(),
	}
	
	// 保存到存储
	accountData := &repository.AccountData{
		Username:      account.Username,
		Password:      account.Password,
		PlayerID:      account.PlayerID,
		PlayerName:    account.PlayerName,
		Email:         "",
		CreateTime:    account.CreateTime,
		LastLoginTime: account.LastLoginTime,
		LoginCount:    0,
		Status:        "active",
	}
	
	if err := as.accountRepo.Save(accountData); err != nil {
		utils.Error("保存账号到存储失败: %v", err)
		return nil, fmt.Errorf("注册失败: %v", err)
	}
	
	// 缓存到内存
	as.accounts[username] = account
	utils.Info("账号注册成功: %s, PlayerID: %s", username, account.PlayerID)
	
	return account, nil
}

// Login 登录
func (as *AccountServer) Login(username, password string) (*Session, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	var account *Account
	
	// 先从内存缓存查找
	account, exists := as.accounts[username]
	if !exists {
		// 从存储加载
		accountData, err := as.accountRepo.GetByUsername(username)
		if err != nil {
			// 账号不存在
			return nil, fmt.Errorf("账号不存在")
		}
		// 从存储数据转换
		account = &Account{
			Username:      accountData.Username,
			Password:      accountData.Password,
			PlayerID:      accountData.PlayerID,
			PlayerName:    accountData.PlayerName,
			CreateTime:    accountData.CreateTime,
			LastLoginTime: accountData.LastLoginTime,
		}
		as.accounts[username] = account
	}
	
	// 验证密码
	if account.Password != hashPassword(password) {
		return nil, fmt.Errorf("密码错误")
	}
	
	// 更新登录时间
	account.LastLoginTime = time.Now()
	
	// 更新存储
	if err := as.accountRepo.UpdateLastLogin(username); err != nil {
		utils.Warn("更新登录时间失败: %v", err)
	}
	
	// 使该玩家之前的token失效（防止多设备同时在线）
	as.invalidateUserTokens(account.PlayerID)
	
	// 生成新token
	token := uuid.New().String()
	session := &Session{
		Token:      token,
		PlayerID:   account.PlayerID,
		Username:   account.Username,
		ExpireTime: time.Now().Add(24 * time.Hour), // 24小时有效
	}
	
	as.tokens[token] = session
	utils.Info("玩家登录成功: %s, PlayerID: %s, Token: %s", username, account.PlayerID, token)
	
	return session, nil
}

// invalidateUserTokens 使指定玩家的所有旧token失效（内部方法，调用时已持有锁）
func (as *AccountServer) invalidateUserTokens(playerID string) {
	for token, session := range as.tokens {
		if session.PlayerID == playerID {
			delete(as.tokens, token)
			utils.Info("使旧token失效: %s", token)
		}
	}
}

// VerifyToken 验证token
func (as *AccountServer) VerifyToken(token string) (*Session, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	session, exists := as.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token不存在")
	}
	
	if time.Now().After(session.ExpireTime) {
		return nil, fmt.Errorf("token已过期")
	}
	
	return session, nil
}

// InvalidateToken 主动使 token 失效（客户端断开连接时调用）
func (as *AccountServer) InvalidateToken(token string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	if session, exists := as.tokens[token]; exists {
		delete(as.tokens, token)
		utils.Info("主动使token失效: %s (玩家: %s)", token, session.Username)
	}
}

// GetTokenCount 获取当前 token 数量（用于监控）
func (as *AccountServer) GetTokenCount() int {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return len(as.tokens)
}

// GetGameServerList 获取区服列表
func (as *AccountServer) GetGameServerList() []*GameServerInfo {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	// 返回副本
	servers := make([]*GameServerInfo, len(as.gameServers))
	copy(servers, as.gameServers)
	return servers
}

// UpdateServerStatus 更新区服状态
func (as *AccountServer) UpdateServerStatus(serverID int, onlineNum int) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	for _, server := range as.gameServers {
		if server.ServerID == serverID {
			server.OnlineNum = onlineNum
			if onlineNum >= server.MaxPlayer {
				server.Status = "full"
			} else {
				server.Status = "online"
			}
			break
		}
	}
}

// hashPassword MD5加密密码
func hashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ========== HTTP 处理器 ==========

// HandleRegister 注册接口
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求数据格式错误")
		return
	}
	
	if req.Username == "" || req.Password == "" {
		sendError(w, "用户名和密码不能为空")
		return
	}
	
	_, err := GetAccountServer().RegisterAccount(req.Username, req.Password)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	
	// 注册成功后自动登录，生成token并返回服务器列表
	session, err := GetAccountServer().Login(req.Username, req.Password)
	if err != nil {
		sendError(w, "注册成功但登录失败: "+err.Error())
		return
	}
	
	servers := GetAccountServer().GetGameServerList()
	
	// 返回token和服务器列表
	sendSuccess(w, map[string]interface{}{
		"token":       session.Token,
		"username":    session.Username,
		"expire_time": session.ExpireTime.Unix(),
		"servers":     servers,
	})
}

// HandleLogin 登录接口
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求数据格式错误")
		return
	}
	
	if req.Username == "" || req.Password == "" {
		sendError(w, "用户名和密码不能为空")
		return
	}
	
	session, err := GetAccountServer().Login(req.Username, req.Password)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	
	servers := GetAccountServer().GetGameServerList()
	
	// 登录成功直接返回token和服务器列表
	sendSuccess(w, map[string]interface{}{
		"token":       session.Token,
		"username":    session.Username,
		"expire_time": session.ExpireTime.Unix(),
		"servers":     servers,
	})
}

// HandleGetServerList 获取区服列表接口（无需登录）
func HandleGetServerList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	servers := GetAccountServer().GetGameServerList()
	
	sendSuccess(w, map[string]interface{}{
		"servers": servers,
	})
}

// sendSuccess 发送成功响应
func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	}
	
	json.NewEncoder(w).Encode(response)
}

// sendError 发送错误响应
func sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"code":    -1,
		"message": message,
		"data":    nil,
	}
	
	json.NewEncoder(w).Encode(response)
}
