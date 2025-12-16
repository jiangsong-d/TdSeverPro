# 塔防服务器 Protobuf 集成指南

## 🔄 从JSON切换到Protobuf

### 为什么使用Protobuf？

✅ **性能更好**: 二进制格式，体积更小，解析更快  
✅ **类型安全**: 强类型，编译时检查  
✅ **向后兼容**: 协议演进更容易  
✅ **跨语言**: 完美配合Unity C#客户端  

## 📦 安装依赖

### 1. 安装 protoc 编译器

**Windows:**
```bash
# 下载并安装
https://github.com/protocolbuffers/protobuf/releases
# 添加到PATH环境变量
```

**Linux/Mac:**
```bash
brew install protobuf
# 或
apt-get install protobuf-compiler
```

### 2. 安装 Go Protobuf 插件

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### 3. 更新服务器依赖

在 `TowerDefenseServer/go.mod` 中已包含：
```go
require (
    google.golang.org/protobuf v1.31.0
)
```

运行：
```bash
cd TowerDefenseServer
go mod download
```

## 🚀 生成 Go 代码

```bash
cd E:\tafang\ConfigTool\Proto
gen_go.bat
```

生成的代码在: `TowerDefenseServer/proto/tower_defense.pb.go`

## 📝 使用示例

### 服务器端发送消息

```go
// 1. 创建消息
resp := &proto.LoginResponse{
    Success:  true,
    PlayerId: "player_123",
    Message:  "登录成功",
}

// 2. 序列化
data, err := proto.Marshal(resp)
if err != nil {
    return err
}

// 3. 包装
gameMsg := &proto.GameMessage{
    Type:    proto.MessageType_MSG_LOGIN,
    Payload: data,
}

// 4. 发送
msgData, _ := proto.Marshal(gameMsg)
session.Send <- msgData
```

### 服务器端接收消息

```go
// 1. 解析外层
gameMsg := &proto.GameMessage{}
if err := proto.Unmarshal(data, gameMsg); err != nil {
    return err
}

// 2. 根据类型解析内层
switch gameMsg.Type {
case proto.MessageType_MSG_LOGIN:
    req := &proto.LoginRequest{}
    proto.Unmarshal(gameMsg.Payload, req)
    handleLogin(req)
}
```

## 🔄 迁移步骤

### 1. 保留现有JSON代码（可选）

如果需要同时支持JSON和Protobuf：

```go
// 检测消息格式
if data[0] == '{' {
    // JSON格式
    handleJSONMessage(data)
} else {
    // Protobuf格式
    handleProtobufMessage(data)
}
```

### 2. 完全切换到Protobuf

建议完全切换，代码更简洁：

- ✅ 删除 `network/protocol.go` 中的JSON结构
- ✅ 使用生成的 `proto/tower_defense.pb.go`
- ✅ 更新 `network/handler.go` 使用protobuf消息

## 🎯 配置说明

### proto文件位置

```
项目结构：
E:\tafang\ConfigTool\Proto\proto\tower_defense.proto  (源文件)
E:\tafang\TowerDefenseServer\proto\                    (Go生成代码)
E:\tafang\Assets\HotUpdate\Network\Proto\              (C#生成代码)
```

### 修改协议

1. 编辑 `tower_defense.proto`
2. 运行 `gen_go.bat` 和 `gen_csharp.bat`
3. 重启服务器和Unity

## ⚡ 性能对比

| 指标 | JSON | Protobuf | 提升 |
|------|------|----------|------|
| 消息大小 | 100% | 30-50% | 2-3倍 |
| 序列化速度 | 100% | 200-300% | 2-3倍 |
| 反序列化速度 | 100% | 300-400% | 3-4倍 |

## 🐛 常见问题

### 1. 生成代码失败

```bash
# 检查protoc版本
protoc --version
# 应该是 libprotoc 3.x 或更高

# 检查Go插件
which protoc-gen-go
```

### 2. 导入错误

```go
// 确保导入路径正确
import pb "towerdefense/proto"
```

### 3. Unity中找不到类型

```csharp
// 确保命名空间正确
using TowerDefense.Proto;
```

## 📚 参考

- [Protobuf Go教程](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [Protobuf语法指南](https://developers.google.com/protocol-buffers/docs/proto3)

---

**注意**: 当前服务器代码仍使用JSON，建议按照上述步骤迁移到Protobuf
