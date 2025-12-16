# 塔防游戏服务器 (TowerDefenseServer)

基于 **Golang + WebSocket** 的轻量级塔防游戏服务器框架，专为配合 Unity 客户端开发。

## 🎮 功能特性

### ✅ 核心功能
- ✅ WebSocket 长连接通信
- ✅ 房间系统（创建、加入、离开）
- ✅ 战斗系统（波次管理、敌人生成、防御塔攻击）
- ✅ 实时状态同步（20帧/秒）
- ✅ 玩家管理（金币、生命值、分数）
- ✅ 防御塔系统（放置、升级、出售）
- ✅ 敌人系统（寻路、血量、速度）
- ✅ 波次系统（自动生成、难度递增）

### 🔧 技术栈
- **语言**: Golang 1.21+
- **WebSocket**: gorilla/websocket
- **UUID**: google/uuid
- **架构**: 单例模式 + 并发安全

## 📁 项目结构

```
TowerDefenseServer/
├── main.go                 # 服务器入口
├── go.mod                  # Go 模块定义
├── config.json             # 配置文件
├── config/                 # 配置管理
│   └── config.go
├── network/                # 网络层
│   ├── websocket.go       # WebSocket 处理
│   ├── session.go         # 会话管理
│   ├── protocol.go        # 通信协议
│   └── handler.go         # 消息处理
├── game/                   # 游戏逻辑
│   ├── player.go          # 玩家
│   ├── room.go            # 房间
│   ├── battle.go          # 战斗
│   ├── tower.go           # 防御塔
│   ├── enemy.go           # 敌人
│   └── wave.go            # 波次
├── logic/                  # 管理器层
│   ├── room_manager.go    # 房间管理器
│   └── battle_manager.go  # 战斗管理器
└── utils/                  # 工具类
    └── logger.go          # 日志工具
```

## 🚀 快速开始

### 1. 安装依赖

```bash
cd TowerDefenseServer
go mod download
```

### 2. 运行服务器

```bash
go run main.go
# 或指定端口
go run main.go -addr=:8080
```

### 3. 服务器将在 `http://localhost:8080` 启动

- WebSocket 端点: `ws://localhost:8080/ws`
- 健康检查: `http://localhost:8080/health`

## 📡 通信协议

### 消息格式

所有消息使用 JSON 格式，统一结构：

```json
{
  "type": 1001,
  "data": { /* 具体数据 */ }
}
```

### 消息类型

| 类型码 | 名称 | 说明 |
|--------|------|------|
| 1000 | Heartbeat | 心跳 |
| 1001 | Login | 登录 |
| 2001 | CreateRoom | 创建房间 |
| 2002 | JoinRoom | 加入房间 |
| 2003 | LeaveRoom | 离开房间 |
| 2004 | RoomInfo | 房间信息 |
| 2005 | StartGame | 开始游戏 |
| 3001 | PlaceTower | 放置防御塔 |
| 3002 | UpgradeTower | 升级防御塔 |
| 3003 | SellTower | 出售防御塔 |
| 3004 | WaveStart | 波次开始 |
| 3005 | WaveComplete | 波次完成 |
| 3006 | GameOver | 游戏结束 |
| 4001 | SyncState | 状态同步 |
| 4002 | SyncEnemy | 敌人同步 |
| 4003 | SyncTower | 防御塔同步 |
| 4004 | SyncDamage | 伤害同步 |
| 9999 | Error | 错误消息 |

### 示例：登录

**客户端发送：**
```json
{
  "type": 1001,
  "data": {
    "player_id": "player_123",
    "player_name": "张三",
    "token": "your_token_here"
  }
}
```

**服务器响应：**
```json
{
  "type": 1001,
  "data": {
    "success": true,
    "player_id": "player_123",
    "message": "登录成功"
  }
}
```

### 示例：创建房间

**客户端发送：**
```json
{
  "type": 2001,
  "data": {
    "room_name": "我的房间",
    "max_player": 4,
    "level_id": 1
  }
}
```

### 示例：放置防御塔

**客户端发送：**
```json
{
  "type": 3001,
  "data": {
    "tower_type": 1,
    "pos_x": 5.0,
    "pos_y": 0.0,
    "pos_z": 5.0
  }
}
```

### 示例：状态同步（服务器推送）

```json
{
  "type": 4001,
  "data": {
    "gold": 150,
    "life": 18,
    "wave_num": 3,
    "enemies": [
      {
        "enemy_id": "enemy_001",
        "type": 1,
        "hp": 30,
        "max_hp": 50,
        "pos_x": 10.5,
        "pos_y": 0.0,
        "pos_z": 8.2,
        "speed": 2.0
      }
    ],
    "towers": [
      {
        "tower_id": "tower_001",
        "type": 1,
        "level": 1,
        "pos_x": 5.0,
        "pos_y": 0.0,
        "pos_z": 5.0,
        "target_id": "enemy_001"
      }
    ]
  }
}
```

## 🎯 游戏流程

1. **连接** → WebSocket 连接到服务器
2. **登录** → 发送玩家信息
3. **创建/加入房间** → 进入游戏房间
4. **开始游戏** → 房主发起开始
5. **游戏进行中**:
   - 放置防御塔
   - 服务器自动生成敌人
   - 防御塔自动攻击敌人
   - 实时同步游戏状态
6. **游戏结束** → 胜利或失败

## ⚙️ 配置说明

编辑 `config.json`:

```json
{
  "server": {
    "port": ":8080",
    "max_players": 1000,
    "room_capacity": 4,
    "heartbeat_interval": 30,
    "session_timeout": 120,
    "tick_rate": 20
  },
  "game": {
    "initial_gold": 100,
    "initial_life": 20,
    "wave_interval": 5.0,
    "enemy_spawn_interval": 1.0
  }
}
```

## 🎨 Unity 客户端集成

### C# WebSocket 连接示例

```csharp
using UnityEngine;
using System;
using NativeWebSocket;

public class GameNetworkManager : MonoBehaviour
{
    private WebSocket websocket;
    
    async void Start()
    {
        websocket = new WebSocket("ws://localhost:8080/ws");
        
        websocket.OnOpen += () =>
        {
            Debug.Log("连接成功");
            SendLogin("player_123", "玩家名");
        };
        
        websocket.OnMessage += (bytes) =>
        {
            var message = System.Text.Encoding.UTF8.GetString(bytes);
            HandleMessage(message);
        };
        
        await websocket.Connect();
    }
    
    void Update()
    {
        #if !UNITY_WEBGL || UNITY_EDITOR
        websocket?.DispatchMessageQueue();
        #endif
    }
    
    async void SendLogin(string playerId, string playerName)
    {
        var msg = new {
            type = 1001,
            data = new {
                player_id = playerId,
                player_name = playerName,
                token = "test_token"
            }
        };
        
        var json = JsonUtility.ToJson(msg);
        await websocket.SendText(json);
    }
}
```

### 推荐 WebSocket 库
- **NativeWebSocket** (推荐): https://github.com/endel/NativeWebSocket
- **WebSocketSharp**: https://github.com/sta/websocket-sharp

## 🔒 安全建议

### 生产环境必须实现：

1. **Token 验证**: 实现真实的用户认证
2. **Origin 检查**: 限制 WebSocket 连接来源
3. **速率限制**: 防止消息洪水攻击
4. **数据验证**: 验证所有客户端输入
5. **加密传输**: 使用 WSS (WebSocket Secure)

## 📊 性能优化

- ✅ 对象池复用（敌人、塔）
- ✅ 增量同步（仅同步变化数据）
- ✅ 空间分区（大规模敌人优化）
- ✅ 定时清理（空房间、断线会话）

## 🛠️ 开发建议

### 扩展配置表

将硬编码数据移到配置文件：
- 防御塔属性 → `tower_config.json`
- 敌人属性 → `enemy_config.json`
- 关卡数据 → `level_config.json`

### 数据持久化

添加数据库支持：
```go
// 推荐使用
- Redis (会话、排行榜)
- MySQL/PostgreSQL (玩家数据)
- MongoDB (游戏记录)
```

### 横向扩展

多服务器架构：
```
LoadBalancer → [Server1, Server2, Server3]
                     ↓
               Redis Pub/Sub
```

## 📝 日志

服务器日志输出到：
- **控制台**: 实时查看
- **server.log**: 文件记录

日志级别：
- `[INFO]` - 一般信息
- `[WARN]` - 警告信息
- `[ERROR]` - 错误信息

## 🐛 常见问题

### 1. 连接失败
- 检查服务器是否运行: `http://localhost:8080/health`
- 检查防火墙设置
- 确认端口未被占用

### 2. 消息未响应
- 查看服务器日志
- 检查消息格式是否正确
- 确认已登录

### 3. 游戏卡顿
- 调低 `tick_rate` (默认20帧)
- 减少同步频率
- 优化敌人数量

## 📈 性能指标

| 指标 | 推荐值 |
|------|--------|
| 同时在线玩家 | <1000 |
| 房间数 | <100 |
| 每房间玩家 | 1-4 |
| 同屏敌人 | <50 |
| 网络延迟 | <100ms |
| 帧率 | 20 FPS |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

## 💡 下一步计划

- [ ] 添加数据库支持
- [ ] 实现排行榜系统
- [ ] 添加好友系统
- [ ] 实现回放功能
- [ ] 支持 HTTPS/WSS
- [ ] Docker 容器化部署
- [ ] 性能监控面板

---

**需要帮助？** 提交 Issue 或联系开发者

**祝你开发愉快！** 🎮
