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
	Host       string `json:"host"`
	Port       int    `json:"port"`
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
		utils.Info("账号服务器初始化完成")
	})
	return accountServer
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
			Host:       "192.168.2.100",
			Port:       8081,
			Status:     "online",
			OnlineNum:  125,
			MaxPlayer:  1000,
			Recommend:  true,
			IsNew:      false,
		},
		{
			ServerID:   2,
			ServerName: "二区-寒冰",
			Host:       "192.168.2.100",
			Port:       8082,
			Status:     "online",
			OnlineNum:  68,
			MaxPlayer:  1000,
			Recommend:  false,
			IsNew:      true,
		},
		{
			ServerID:   3,
			ServerName: "三区-雷霆",
			Host:       "192.168.2.100",
			Port:       8083,
			Status:     "maintain",
			OnlineNum:  0,
			MaxPlayer:  1000,
			Recommend:  false,
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
			// 账号不存在，自动注册（开发环境）
			account = &Account{
				Username:      username,
				Password:      hashPassword(password),
				PlayerID:      uuid.New().String(),
				PlayerName:    username,
				CreateTime:    time.Now(),
				LastLoginTime: time.Now(),
			}
			
			// 保存到存储
			newAccountData := &repository.AccountData{
				Username:      account.Username,
				Password:      account.Password,
				PlayerID:      account.PlayerID,
				PlayerName:    account.PlayerName,
				Email:         "",
				CreateTime:    account.CreateTime,
				LastLoginTime: account.LastLoginTime,
				LoginCount:    1,
				Status:        "active",
			}
			
			if err := as.accountRepo.Save(newAccountData); err != nil {
				utils.Error("保存账号失败: %v", err)
				return nil, fmt.Errorf("登录失败: %v", err)
			}
			
			as.accounts[username] = account
			utils.Info("自动注册新账号: %s", username)
		} else {
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
	
	// 生成token
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
	
	account, err := GetAccountServer().RegisterAccount(req.Username, req.Password)
	if err != nil {
		sendError(w, err.Error())
		return
	}
	
	sendSuccess(w, map[string]interface{}{
		"player_id":   account.PlayerID,
		"player_name": account.PlayerName,
		"message":     "注册成功",
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
	
	sendSuccess(w, map[string]interface{}{
		"token":       session.Token,
		"player_id":   session.PlayerID,
		"username":    session.Username,
		"expire_time": session.ExpireTime.Unix(),
		"message":     "登录成功",
	})
}

// HandleGetServerList 获取区服列表接口
func HandleGetServerList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 验证token
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
	} else {
		// 去掉 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}
	
	if token == "" {
		utils.Error("获取服务器列表失败: 缺少token")
		sendError(w, "缺少token")
		return
	}
	
	utils.Info("验证token: %s", token)
	
	session, err := GetAccountServer().VerifyToken(token)
	if err != nil {
		utils.Error("token验证失败: %v", err)
		sendError(w, err.Error())
		return
	}
	
	utils.Info("token验证成功，玩家: %s", session.Username)
	
	servers := GetAccountServer().GetGameServerList()
	
	sendSuccess(w, map[string]interface{}{
		"player_id": session.PlayerID,
		"servers":   servers,
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
