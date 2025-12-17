package storage

// IStorage 统一存储接口
// 支持 TXT、MySQL、Redis 等多种存储方式
type IStorage interface {
	// 初始化存储
	Init(config map[string]interface{}) error
	
	// 关闭存储连接
	Close() error
	
	// 保存数据（通用）
	Save(table string, key string, data interface{}) error
	
	// 获取数据（通用）
	Get(table string, key string, result interface{}) error
	
	// 删除数据
	Delete(table string, key string) error
	
	// 查询所有数据
	GetAll(table string) ([]interface{}, error)
	
	// 条件查询（用于数据库）
	Query(table string, condition map[string]interface{}) ([]interface{}, error)
	
	// 批量保存
	SaveBatch(table string, items map[string]interface{}) error
	
	// 检查是否存在
	Exists(table string, key string) (bool, error)
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeTXT   StorageType = "txt"
	StorageTypeMySQL StorageType = "mysql"
	StorageTypeRedis StorageType = "redis"
)

// StorageConfig 存储配置
type StorageConfig struct {
	Type     StorageType            `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}
