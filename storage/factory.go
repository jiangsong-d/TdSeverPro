package storage

import (
	"fmt"
	"sync"
	"towerdefense/utils"
)

var (
	globalStorage IStorage
	storageOnce   sync.Once
)

// InitStorage 初始化存储（根据配置选择存储方式）
func InitStorage(storageType StorageType, config map[string]interface{}) error {
	var err error
	storageOnce.Do(func() {
		var storage IStorage
		
		switch storageType {
		case StorageTypeTXT:
			storage = NewTxtStorage()
		case StorageTypeMySQL:
			storage = NewMySQLStorage()
		case StorageTypeRedis:
			storage = NewRedisStorage()
		default:
			err = fmt.Errorf("不支持的存储类型: %s", storageType)
			return
		}
		
		if err = storage.Init(config); err != nil {
			return
		}
		
		globalStorage = storage
		utils.Info("存储层初始化完成，类型: %s", storageType)
	})
	
	return err
}

// GetStorage 获取全局存储实例
func GetStorage() IStorage {
	if globalStorage == nil {
		// 默认使用TXT存储
		_ = InitStorage(StorageTypeTXT, map[string]interface{}{
			"data_dir": "./data",
		})
	}
	return globalStorage
}

// CloseStorage 关闭存储
func CloseStorage() error {
	if globalStorage != nil {
		return globalStorage.Close()
	}
	return nil
}
