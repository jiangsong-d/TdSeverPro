package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"towerdefense/utils"
)

// TxtStorage TXT文件存储实现
// 使用 JSON 格式存储，便于后期迁移到数据库
type TxtStorage struct {
	dataDir string // 数据目录
	mu      sync.RWMutex
}

// NewTxtStorage 创建TXT存储
func NewTxtStorage() *TxtStorage {
	return &TxtStorage{}
}

// Init 初始化存储
func (ts *TxtStorage) Init(config map[string]interface{}) error {
	// 获取数据目录
	dataDir := "./data"
	if dir, ok := config["data_dir"].(string); ok {
		dataDir = dir
	}
	
	ts.dataDir = dataDir
	
	// 创建数据目录
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}
	
	utils.Info("TXT存储初始化完成，数据目录: %s", dataDir)
	return nil
}

// Close 关闭存储
func (ts *TxtStorage) Close() error {
	utils.Info("TXT存储关闭")
	return nil
}

// getTablePath 获取表文件路径
func (ts *TxtStorage) getTablePath(table string) string {
	return filepath.Join(ts.dataDir, table)
}

// getFilePath 获取数据文件路径
func (ts *TxtStorage) getFilePath(table, key string) string {
	tablePath := ts.getTablePath(table)
	return filepath.Join(tablePath, fmt.Sprintf("%s.json", key))
}

// ensureTableDir 确保表目录存在
func (ts *TxtStorage) ensureTableDir(table string) error {
	tablePath := ts.getTablePath(table)
	return os.MkdirAll(tablePath, 0755)
}

// Save 保存数据
func (ts *TxtStorage) Save(table string, key string, data interface{}) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	// 确保表目录存在
	if err := ts.ensureTableDir(table); err != nil {
		return err
	}
	
	// 序列化为JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	
	// 写入文件
	filePath := ts.getFilePath(table, key)
	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	return nil
}

// Get 获取数据
func (ts *TxtStorage) Get(table string, key string, result interface{}) error {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	// 读取文件
	filePath := ts.getFilePath(table, key)
	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("数据不存在: %s/%s", table, key)
		}
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 反序列化
	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("JSON反序列化失败: %v", err)
	}
	
	return nil
}

// Delete 删除数据
func (ts *TxtStorage) Delete(table string, key string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	filePath := ts.getFilePath(table, key)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为删除成功
		}
		return fmt.Errorf("删除文件失败: %v", err)
	}
	
	return nil
}

// GetAll 获取所有数据
func (ts *TxtStorage) GetAll(table string) ([]interface{}, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	tablePath := ts.getTablePath(table)
	
	// 检查目录是否存在
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return []interface{}{}, nil
	}
	
	// 读取目录下所有文件
	files, err := ioutil.ReadDir(tablePath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}
	
	results := make([]interface{}, 0)
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		// 读取文件内容
		filePath := filepath.Join(tablePath, file.Name())
		jsonData, err := ioutil.ReadFile(filePath)
		if err != nil {
			utils.Warn("读取文件失败: %s, %v", filePath, err)
			continue
		}
		
		// 反序列化为 map
		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			utils.Warn("JSON反序列化失败: %s, %v", filePath, err)
			continue
		}
		
		results = append(results, data)
	}
	
	return results, nil
}

// Query 条件查询（简单实现，数据库版本会更高效）
func (ts *TxtStorage) Query(table string, condition map[string]interface{}) ([]interface{}, error) {
	// 获取所有数据
	allData, err := ts.GetAll(table)
	if err != nil {
		return nil, err
	}
	
	// 如果没有条件，返回所有数据
	if len(condition) == 0 {
		return allData, nil
	}
	
	// 过滤数据
	results := make([]interface{}, 0)
	for _, item := range allData {
		dataMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		// 检查是否匹配所有条件
		matched := true
		for key, value := range condition {
			if dataMap[key] != value {
				matched = false
				break
			}
		}
		
		if matched {
			results = append(results, item)
		}
	}
	
	return results, nil
}

// SaveBatch 批量保存
func (ts *TxtStorage) SaveBatch(table string, items map[string]interface{}) error {
	for key, data := range items {
		if err := ts.Save(table, key, data); err != nil {
			return err
		}
	}
	return nil
}

// Exists 检查是否存在
func (ts *TxtStorage) Exists(table string, key string) (bool, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	filePath := ts.getFilePath(table, key)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}
