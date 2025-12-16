# Unity 客户端 WebSocket 集成指南

## 安装 WebSocket 库

### 方法1: 使用 NativeWebSocket (推荐)

```bash
# 在 Unity Package Manager 中添加
https://github.com/endel/NativeWebSocket.git#upm
```

### 方法2: 使用 WebSocketSharp

下载并导入到 Unity 项目的 `Plugins` 文件夹

## C# 网络管理器完整示例

创建文件: `Assets/HotUpdate/Network/GameNetworkManager.cs`

```csharp
using UnityEngine;
using System;
using System.Collections.Generic;
using NativeWebSocket;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

/// <summary>
/// 游戏网络管理器
/// 负责与塔防服务器的 WebSocket 通信
/// </summary>
public class GameNetworkManager : MonoSingleton<GameNetworkManager>
{
    private WebSocket websocket;
    private string serverUrl = "ws://localhost:8080/ws";
    
    // 消息回调
    private Dictionary<int, Action<JObject>> messageHandlers = new Dictionary<int, Action<JObject>>();
    
    // 连接状态
    public bool IsConnected { get; private set; }
    
    protected override void Init()
    {
        RegisterMessageHandlers();
    }
    
    /// <summary>
    /// 连接服务器
    /// </summary>
    public async void Connect(string url = null)
    {
        if (!string.IsNullOrEmpty(url))
            serverUrl = url;
            
        websocket = new WebSocket(serverUrl);
        
        websocket.OnOpen += () =>
        {
            IsConnected = true;
            LogUtlis.Info("服务器连接成功");
            OnConnected();
        };
        
        websocket.OnMessage += (bytes) =>
        {
            var message = System.Text.Encoding.UTF8.GetString(bytes);
            HandleMessage(message);
        };
        
        websocket.OnError += (e) =>
        {
            LogUtlis.Error($"WebSocket 错误: {e}");
        };
        
        websocket.OnClose += (e) =>
        {
            IsConnected = false;
            LogUtlis.Warn($"连接关闭: {e}");
            OnDisconnected();
        };
        
        await websocket.Connect();
    }
    
    /// <summary>
    /// 断开连接
    /// </summary>
    public async void Disconnect()
    {
        if (websocket != null && websocket.State == WebSocketState.Open)
        {
            await websocket.Close();
        }
    }
    
    private void Update()
    {
        #if !UNITY_WEBGL || UNITY_EDITOR
        websocket?.DispatchMessageQueue();
        #endif
    }
    
    private void OnDestroy()
    {
        Disconnect();
    }
    
    /// <summary>
    /// 注册消息处理器
    /// </summary>
    private void RegisterMessageHandlers()
    {
        messageHandlers[1001] = OnLoginResponse;
        messageHandlers[2001] = OnCreateRoomResponse;
        messageHandlers[2002] = OnJoinRoomResponse;
        messageHandlers[2004] = OnRoomInfo;
        messageHandlers[2005] = OnStartGame;
        messageHandlers[3001] = OnPlaceTowerResponse;
        messageHandlers[3004] = OnWaveStart;
        messageHandlers[3005] = OnWaveComplete;
        messageHandlers[3006] = OnGameOver;
        messageHandlers[4001] = OnSyncState;
        messageHandlers[4004] = OnSyncDamage;
        messageHandlers[9999] = OnError;
    }
    
    /// <summary>
    /// 处理消息
    /// </summary>
    private void HandleMessage(string message)
    {
        try
        {
            var json = JObject.Parse(message);
            int msgType = json["type"].Value<int>();
            var data = json["data"] as JObject;
            
            if (messageHandlers.ContainsKey(msgType))
            {
                messageHandlers[msgType]?.Invoke(data);
            }
            else
            {
                LogUtlis.Warn($"未处理的消息类型: {msgType}");
            }
        }
        catch (Exception e)
        {
            LogUtlis.Error($"消息解析失败: {e.Message}");
        }
    }
    
    /// <summary>
    /// 发送消息
    /// </summary>
    private async void SendMessage(int msgType, object data)
    {
        if (!IsConnected)
        {
            LogUtlis.Error("未连接到服务器");
            return;
        }
        
        var msg = new
        {
            type = msgType,
            data = data
        };
        
        var json = JsonConvert.SerializeObject(msg);
        await websocket.SendText(json);
    }
    
    // ==================== API 方法 ====================
    
    /// <summary>
    /// 登录
    /// </summary>
    public void Login(string playerId, string playerName)
    {
        SendMessage(1001, new
        {
            player_id = playerId,
            player_name = playerName,
            token = "test_token"
        });
    }
    
    /// <summary>
    /// 创建房间
    /// </summary>
    public void CreateRoom(string roomName, int maxPlayer, int levelId)
    {
        SendMessage(2001, new
        {
            room_name = roomName,
            max_player = maxPlayer,
            level_id = levelId
        });
    }
    
    /// <summary>
    /// 加入房间
    /// </summary>
    public void JoinRoom(string roomId)
    {
        SendMessage(2002, new
        {
            room_id = roomId
        });
    }
    
    /// <summary>
    /// 开始游戏
    /// </summary>
    public void StartGame()
    {
        SendMessage(2005, new { });
    }
    
    /// <summary>
    /// 放置防御塔
    /// </summary>
    public void PlaceTower(int towerType, Vector3 position)
    {
        SendMessage(3001, new
        {
            tower_type = towerType,
            pos_x = position.x,
            pos_y = position.y,
            pos_z = position.z
        });
    }
    
    /// <summary>
    /// 升级防御塔
    /// </summary>
    public void UpgradeTower(string towerId)
    {
        SendMessage(3002, new
        {
            tower_id = towerId
        });
    }
    
    /// <summary>
    /// 出售防御塔
    /// </summary>
    public void SellTower(string towerId)
    {
        SendMessage(3003, new
        {
            tower_id = towerId
        });
    }
    
    // ==================== 消息处理回调 ====================
    
    private void OnConnected()
    {
        // 自动登录
        string playerId = SystemInfo.deviceUniqueIdentifier;
        string playerName = "玩家" + UnityEngine.Random.Range(1000, 9999);
        Login(playerId, playerName);
    }
    
    private void OnDisconnected()
    {
        // 处理断线逻辑
        EngineEventManager.Instance.SendEvent(EventID.NETWORK_DISCONNECTED);
    }
    
    private void OnLoginResponse(JObject data)
    {
        bool success = data["success"].Value<bool>();
        if (success)
        {
            LogUtlis.Info("登录成功");
            EngineEventManager.Instance.SendEvent(EventID.LOGIN_SUCCESS);
        }
        else
        {
            LogUtlis.Error($"登录失败: {data["message"]}");
        }
    }
    
    private void OnCreateRoomResponse(JObject data)
    {
        bool success = data["success"].Value<bool>();
        if (success)
        {
            string roomId = data["room_id"].Value<string>();
            LogUtlis.Info($"房间创建成功: {roomId}");
            // 通知 UI
        }
    }
    
    private void OnJoinRoomResponse(JObject data)
    {
        bool success = data["success"].Value<bool>();
        if (success)
        {
            // 进入房间界面
            LogUtlis.Info("加入房间成功");
        }
    }
    
    private void OnRoomInfo(JObject data)
    {
        // 更新房间信息 UI
        var players = data["players"];
        LogUtlis.Info($"房间信息更新，玩家数: {players.Count()}");
    }
    
    private void OnStartGame(JObject data)
    {
        // 开始游戏，切换到战斗场景
        LogUtlis.Info("游戏开始");
        LoadingManager.Instance.SwitchScene(LoadSceneType.Battle);
    }
    
    private void OnPlaceTowerResponse(JObject data)
    {
        bool success = data["success"].Value<bool>();
        if (success)
        {
            string towerId = data["tower_id"].Value<string>();
            int gold = data["gold"].Value<int>();
            // 更新 UI 金币显示
            LogUtlis.Info($"防御塔放置成功，剩余金币: {gold}");
        }
    }
    
    private void OnWaveStart(JObject data)
    {
        int waveNum = data["wave_num"].Value<int>();
        LogUtlis.Info($"第 {waveNum} 波开始");
        // 显示波次提示
    }
    
    private void OnWaveComplete(JObject data)
    {
        int waveNum = data["wave_num"].Value<int>();
        int reward = data["reward"].Value<int>();
        LogUtlis.Info($"第 {waveNum} 波完成，奖励: {reward}");
    }
    
    private void OnGameOver(JObject data)
    {
        bool isVictory = data["is_victory"].Value<bool>();
        int score = data["score"].Value<int>();
        LogUtlis.Info($"游戏结束，{(isVictory ? "胜利" : "失败")}，得分: {score}");
        // 显示结算界面
    }
    
    private void OnSyncState(JObject data)
    {
        // 同步游戏状态
        int gold = data["gold"].Value<int>();
        int life = data["life"].Value<int>();
        int waveNum = data["wave_num"].Value<int>();
        
        var enemies = data["enemies"];
        var towers = data["towers"];
        
        // 更新游戏对象位置和状态
        UpdateEnemies(enemies);
        UpdateTowers(towers);
    }
    
    private void OnSyncDamage(JObject data)
    {
        string towerId = data["tower_id"].Value<string>();
        string enemyId = data["enemy_id"].Value<string>();
        int damage = data["damage"].Value<int>();
        bool isCrit = data["is_crit"].Value<bool>();
        
        // 播放攻击特效
        LogUtlis.Info($"塔 {towerId} 攻击敌人 {enemyId}，伤害: {damage}");
    }
    
    private void OnError(JObject data)
    {
        string message = data["message"].Value<string>();
        LogUtlis.Error($"服务器错误: {message}");
    }
    
    // ==================== 游戏对象更新 ====================
    
    private void UpdateEnemies(JToken enemies)
    {
        foreach (var enemy in enemies)
        {
            string enemyId = enemy["enemy_id"].Value<string>();
            int hp = enemy["hp"].Value<int>();
            float posX = enemy["pos_x"].Value<float>();
            float posY = enemy["pos_y"].Value<float>();
            float posZ = enemy["pos_z"].Value<float>();
            
            // 更新或创建敌人对象
            // TODO: 实现敌人管理逻辑
        }
    }
    
    private void UpdateTowers(JToken towers)
    {
        foreach (var tower in towers)
        {
            string towerId = tower["tower_id"].Value<string>();
            string targetId = tower["target_id"].Value<string>();
            
            // 更新防御塔目标
            // TODO: 实现防御塔管理逻辑
        }
    }
}
```

## 使用示例

### 在登录场景

```csharp
public class LoginScene : BaseScene
{
    public override void OnCreate()
    {
        // 连接服务器
        GameNetworkManager.Instance.Connect("ws://localhost:8080/ws");
    }
}
```

### 在主城场景创建房间

```csharp
public void OnCreateRoomButtonClick()
{
    GameNetworkManager.Instance.CreateRoom("我的房间", 4, 1);
}
```

### 在战斗场景放置防御塔

```csharp
public void OnPlaceTower(int towerType, Vector3 position)
{
    GameNetworkManager.Instance.PlaceTower(towerType, position);
}
```

## 事件ID定义

在 `EventID.cs` 中添加：

```csharp
public static class EventID
{
    public const string NETWORK_CONNECTED = "network_connected";
    public const string NETWORK_DISCONNECTED = "network_disconnected";
    public const string LOGIN_SUCCESS = "login_success";
    public const string ROOM_CREATED = "room_created";
    public const string GAME_START = "game_start";
    public const string WAVE_START = "wave_start";
    public const string GAME_OVER = "game_over";
}
```

## 注意事项

1. **安装 Newtonsoft.Json**: Unity Package Manager → Add package by name → `com.unity.nuget.newtonsoft-json`

2. **WebGL 平台**: WebSocket 在 WebGL 上的行为不同，需要特殊处理

3. **断线重连**: 建议实现自动重连机制

4. **消息队列**: 对于高频消息（如位置同步），考虑使用插值平滑

5. **安全性**: 生产环境使用 HTTPS/WSS 加密连接
