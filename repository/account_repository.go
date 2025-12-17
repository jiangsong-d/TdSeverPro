package repository

import (
	"fmt"
	"towerdefense/storage"
	"time"
)

// AccountData 账号数据模型（对应数据库表结构）
type AccountData struct {
	Username      string    `json:"username"`
	Password      string    `json:"password"`       // MD5加密
	PlayerID      string    `json:"player_id"`
	PlayerName    string    `json:"player_name"`
	Email         string    `json:"email"`
	CreateTime    time.Time `json:"create_time"`
	LastLoginTime time.Time `json:"last_login_time"`
	LoginCount    int       `json:"login_count"`
	Status        string    `json:"status"`         // active, banned, deleted
}

const TableAccount = "accounts"

// AccountRepository 账号仓储
type AccountRepository struct {
	storage storage.IStorage
}

// NewAccountRepository 创建账号仓储
func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		storage: storage.GetStorage(),
	}
}

// Save 保存账号
func (ar *AccountRepository) Save(account *AccountData) error {
	return ar.storage.Save(TableAccount, account.Username, account)
}

// GetByUsername 根据用户名获取账号
func (ar *AccountRepository) GetByUsername(username string) (*AccountData, error) {
	var account AccountData
	err := ar.storage.Get(TableAccount, username, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByPlayerID 根据玩家ID获取账号
func (ar *AccountRepository) GetByPlayerID(playerID string) (*AccountData, error) {
	// 查询所有账号，找到匹配的
	results, err := ar.storage.Query(TableAccount, map[string]interface{}{
		"player_id": playerID,
	})
	if err != nil {
		return nil, err
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("账号不存在: player_id=%s", playerID)
	}
	
	// 转换为 AccountData
	var account AccountData
	// 注意：这里需要根据实际返回的类型进行转换
	// TXT存储返回的是 map[string]interface{}
	return &account, nil
}

// Delete 删除账号
func (ar *AccountRepository) Delete(username string) error {
	return ar.storage.Delete(TableAccount, username)
}

// Exists 检查账号是否存在
func (ar *AccountRepository) Exists(username string) (bool, error) {
	return ar.storage.Exists(TableAccount, username)
}

// GetAll 获取所有账号
func (ar *AccountRepository) GetAll() ([]*AccountData, error) {
	results, err := ar.storage.GetAll(TableAccount)
	if err != nil {
		return nil, err
	}
	
	accounts := make([]*AccountData, 0, len(results))
	for _, result := range results {
		// 这里需要类型转换逻辑
		// 暂时返回空列表
		_ = result
	}
	
	return accounts, nil
}

// UpdateLastLogin 更新最后登录时间
func (ar *AccountRepository) UpdateLastLogin(username string) error {
	account, err := ar.GetByUsername(username)
	if err != nil {
		return err
	}
	
	account.LastLoginTime = time.Now()
	account.LoginCount++
	
	return ar.Save(account)
}
