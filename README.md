# Chat Backend
毕设项目
一个基于Go语言开发的实时聊天室后端服务，提供完整的即时通讯功能，包括私聊、群聊、好友管理、群组管理等核心功能。

## 功能特性

- 用户认证与授权
  - 用户注册与登录
  - JWT Token 认证
  - Token 自动刷新

- 好友管理
  - 添加好友（需对方同意）
  - 好友列表查询
  - 删除好友
  - 用户搜索

- 群组管理
  - 创建群组
  - 搜索群组
  - 申请加入群组
  - 审批入群申请
  - 移除群成员
  - 转让群主
  - 退出群组
  - 解散群组

- 实时消息
  - WebSocket 实时通信
  - 私聊消息
  - 群聊消息
  - 消息历史查询
  - 在线用户查询
  - 消息已读回执
  - 用户上线/下线状态通知

- 数据持久化
  - PostgreSQL 数据库
  - Redis 缓存
  - GORM 数据库操作

## 技术栈

- Go 1.25
- Echo - Web 框架
- PostgreSQL - 数据库
- Redis - 缓存
- GORM - ORM 框架
- WebSocket - 实时通信
- JWT - 身份认证
- Zap - 日志记录

## 快速开始

### 环境要求

- Go 1.25 或更高版本
- PostgreSQL 数据库
- Redis 服务

### 安装依赖

```bash
go mod download
```

### 配置说明

复制并编辑配置文件：

```bash
cp configs/config.yaml.example configs/config.yaml
```

修改 `configs/config.yaml` 中的配置项：

```yaml
server:
  port: 8080
  host: "localhost"

database:
  host: "your-database-host"
  port: 5432
  user: "your-database-user"
  password: "your-database-password"
  dbname: "your-database-name"

redis:
  host: "your-redis-host"
  port: 6379
  username: ""
  password: "your-redis-password"
  db: 0

jwt:
  secret: "your-secret-key-here"
  accessExpiry: 24
  refreshExpiry: 168
```

### 数据库迁移

首次运行需要执行数据库迁移：

```bash
go run main.go --migrate
```

如果需要重置数据库（删除所有表并重新创建）：

```bash
go run main.go --reset-db
```

### 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 构建项目

```bash
go build -o chat_backend
```

## API 文档

> **注意**：所有 API 接口都应用了限流保护，详见下方「限流功能」章节。

### 认证相关 🚦 (10次/分钟)

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新 Token

### 用户相关 🚦 (60次/分钟)

- `GET /api/v1/user/me` - 获取当前用户信息
- `GET /api/v1/user/search` - 搜索用户
- `GET /api/v1/user/friend` - 获取好友列表
- `POST /api/v1/user/friend` - 添加好友
- `PUT /api/v1/user/friend/:id` - 处理好友请求
- `DELETE /api/v1/user/friend/:id` - 删除好友

### 群组相关 🚦 (60次/分钟)

- `POST /api/v1/group` - 创建群组
- `GET /api/v1/group` - 获取群组列表
- `GET /api/v1/group/:id` - 获取群组详情
- `GET /api/v1/group/search` - 搜索群组
- `POST /api/v1/group/:id/request-join` - 申请加入群组
- `GET /api/v1/group/join-requests` - 获取待审核的入群请求
- `POST /api/v1/group/:id/join-requests/:user_id/approve` - 审批入群请求
- `POST /api/v1/group/:id/leave` - 退出群组
- `DELETE /api/v1/group/:id` - 解散群组
- `PUT /api/v1/group/:id/transfer` - 转让群组
- `DELETE /api/v1/group/:group_id/member/:user_id` - 移除群组成员

### 消息相关 🚦 (60次/分钟)

- `GET /api/v1/message/conversations` - 获取会话列表
- `GET /api/v1/message/private` - 获取私聊消息记录
- `GET /api/v1/message/group/:id` - 获取群聊消息记录

### WebSocket

- `GET /ws` - WebSocket 连接（需要 JWT 认证）
- `GET /api/v1/ws/online` - 获取在线用户列表
- `GET /api/v1/ws/online/:id` - 查询用户是否在线

#### WebSocket 心跳机制

项目使用标准的 WebSocket ping/pong 控制帧实现心跳机制，确保连接的活跃性和可靠性。

**实现细节：**

- **心跳间隔**：每 30 秒发送一次 ping 控制帧
- **超时控制**：使用 `context.WithTimeout` 设置 10 秒超时时间
- **完整往返验证**：`Ping()` 方法会等待完整的 ping/pong 往返
  - 发送 ping 控制帧
  - 等待对方接收 ping
  - 等待对方回复 pong
  - 接收 pong 响应
- **错误处理**：如果 `Ping()` 返回错误，说明连接可能已断开，会立即关闭连接
- **自动回复**：浏览器会自动回复 pong，无需前端手动处理

**代码实现位置：**[`internal/websocket/connection.go`](internal/websocket/connection.go:174-183)

**日志输出：**
- Ping 成功：`INFO Ping sent successfully user_id=<用户ID>`
- Ping 失败：`ERROR Ping failed, connection may be closed user_id=<用户ID> error=<错误信息>`

### 其他

- `GET /` - 服务欢迎信息
- `GET /health` - 健康检查
- `GET /api/v1/hello` - Hello World 测试接口

## 项目结构

```
chat_backend/
├── internal/
│   ├── config/          # 配置管理
│   ├── dao/             # 数据访问对象
│   ├── database/        # 数据库初始化和迁移
│   ├── dto/             # 数据传输对象
│   ├── errors/          # 错误处理
│   ├── global/          # 全局变量和常量
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── response/        # 统一响应格式
│   ├── router/          # 路由配置
│   ├── service/         # 业务逻辑
│   └── websocket/       # WebSocket 处理
├── pkg/
│   ├── env/             # 环境变量
│   ├── logger/          # 日志工具
│   └── utils/           # 工具函数
├── configs/             # 配置文件
├── main.go              # 程序入口
├── go.mod               # Go 模块文件
└── go.sum               # 依赖锁定文件
```

## Redis 缓存设计

### 缓存架构

项目采用分层缓存架构，包含基础 Redis 封装层和业务缓存管理器层：

#### 1. 基础封装层 ([`internal/cache/redis.go`](internal/cache/redis.go))

[`RedisClient`](internal/cache/redis.go:13) 是对 `github.com/redis/go-redis/v9` 的封装，提供了多种数据结构的操作：

| 数据结构 | 操作类型 | 主要用途 |
|---------|---------|---------|
| **String** | [`StringOperations`](internal/cache/redis.go:30) | 存储单个值、JSON 序列化、过期时间控制 |
| **Hash** | [`HashOperations`](internal/cache/redis.go:98) | 存储对象字段、批量字段操作 |
| **List** | [`ListOperations`](internal/cache/redis.go:166) | 消息队列、列表数据 |
| **Set** | [`SetOperations`](internal/cache/redis.go:216) | 去重集合、好友列表 |
| **Sorted Set** | [`SortedSetOperations`](internal/cache/redis.go:251) | 排序集合、在线状态（按时间戳排序） |
| **Pipeline** | [`PipelineOperations`](internal/cache/redis.go:311) | 批量操作、性能优化 |
| **Transaction** | [`TransactionOperations`](internal/cache/redis.go:331) | 事务操作 |

#### 2. 缓存键设计

使用 [`CacheKey`](internal/cache/redis.go:351) 生成器统一管理缓存键格式：

```go
KeyPrefixRefreshToken    = "refresh_token"
KeyPrefixOnlineUsers     = "online_users"
KeyPrefixUser            = "user"
KeyPrefixUsernameToID    = "username_to_id"
KeyPrefixMessage         = "message"
KeyPrefixMessagesPrivate = "messages:private"
KeyPrefixMessagesGroup   = "messages:group"
KeyPrefixConversations   = "conversations"
KeyPrefixFriends         = "friends"
KeyPrefixFriendRequests  = "friend_requests"
KeyPrefixRateLimit       = "rate_limit"
```

### 六大缓存管理器

#### 1. 消息缓存管理器 ([`MessageCacheManager`](internal/cache/message_cache.go:39))

**数据结构：**
- 使用 Redis List 存储消息 ID 列表（最新消息在左侧）
- 使用 Redis String 存储消息详情（JSON 格式）

**TTL 配置：**
- `MessageTTL = 24 * time.Hour`（24小时）
- `MessageCacheLimit = 100`（每个会话最多缓存 100 条消息）

**主要功能：**
- [`CachePrivateMessage`](internal/cache/message_cache.go:52) - 缓存私聊消息
- [`CacheGroupMessage`](internal/cache/message_cache.go:80) - 缓存群聊消息
- [`GetCachedPrivateMessages`](internal/cache/message_cache.go:120) - 获取私聊消息
- [`GetCachedGroupMessages`](internal/cache/message_cache.go:145) - 获取群聊消息
- [`BatchGetMessageDetails`](internal/cache/message_cache.go:186) - 批量获取消息详情（使用 Pipeline）

#### 2. 会话缓存管理器 ([`ConversationCacheManager`](internal/cache/conversation_cache.go:35))

**数据结构：**
- 使用 Redis Hash 存储会话列表
- 字段格式：`private:{userID}` 或 `group:{groupID}`

**TTL 配置：**
- `ConversationTTL = 5 * time.Minute`（5分钟）

**主要功能：**
- [`SetPrivateConversation`](internal/cache/conversation_cache.go:48) - 缓存私聊会话
- [`SetGroupConversation`](internal/cache/conversation_cache.go:72) - 缓存群聊会话
- [`GetAllConversations`](internal/cache/conversation_cache.go:198) - 获取所有会话

#### 3. 好友缓存管理器 ([`FriendCacheManager`](internal/cache/friend_cache.go:27))

**数据结构：**
- 好友列表使用 Redis Set（自动去重）
- 好友申请使用 Redis List

**TTL 配置：**
- `FriendListTTL = 30 * time.Minute`（30分钟）
- `FriendRequestTTL = 5 * time.Minute`（5分钟）

**主要功能：**
- [`AddFriend`](internal/cache/friend_cache.go:40) / [`RemoveFriend`](internal/cache/friend_cache.go:58) - 添加/移除好友
- [`IsFriend`](internal/cache/friend_cache.go:70) - 检查是否是好友
- [`GetFriendList`](internal/cache/friend_cache.go:82) - 获取好友列表
- [`AddFriendRequest`](internal/cache/friend_cache.go:183) - 添加好友申请

#### 4. 用户缓存管理器 ([`UserCacheManager`](internal/cache/user_cache.go:28))

**数据结构：**
- 使用 Redis String 存储用户信息（JSON 格式）
- 使用 Redis String 存储用户名到 ID 的映射

**TTL 配置：**
- `UserInfoTTL = 1 * time.Hour`（1小时）
- `UsernameToIDTTL = 1 * time.Hour`（1小时）

**主要功能：**
- [`SetUserInfo`](internal/cache/user_cache.go:41) / [`GetUserInfo`](internal/cache/user_cache.go:59) - 缓存/获取用户信息
- [`SetUsernameToID`](internal/cache/user_cache.go:131) / [`GetUserIDByUsername`](internal/cache/user_cache.go:143) - 用户名映射
- [`BatchGetUserInfo`](internal/cache/user_cache.go:75) - 批量获取用户信息
- [`GetOrLoadUserInfo`](internal/cache/user_cache.go:227) - Cache-Aside 模式获取

#### 5. 会话管理器 ([`SessionManager`](internal/cache/session.go:34))

**数据结构：**
- 使用 Redis Hash 存储多个设备的 Refresh Token
- 每个用户最多支持 `MaxDevicesPerUser = 5` 个设备

**TTL 配置：**
- `DefaultRefreshTokenTTL = 168 * time.Hour`（7天）

**主要功能：**
- [`StoreRefreshToken`](internal/cache/session.go:54) - 存储 Refresh Token
- [`ValidateRefreshToken`](internal/cache/session.go:105) - 验证 Token
- [`RevokeRefreshToken`](internal/cache/session.go:137) - 撤销指定 Token
- [`RevokeAllUserSessions`](internal/cache/session.go:167) - 撤销用户所有会话
- [`cleanupOldSessions`](internal/cache/session.go:207) - 自动清理旧会话（超过设备数量限制）

#### 6. 在线状态管理器 ([`OnlineStatusManager`](internal/cache/online_status.go:20))

**数据结构：**
- 使用 Redis Sorted Set 存储在线用户
- Score 为最后心跳时间戳（Unix 时间戳）

**TTL 配置：**
- `OnlineStatusTTL = 60 * time.Second`（60秒）
- `OnlineHeartbeatInterval = 30 * time.Second`（30秒心跳间隔）

**主要功能：**
- [`SetOnline`](internal/cache/online_status.go:33) / [`SetOffline`](internal/cache/online_status.go:50) - 设置在线/离线状态
- [`Heartbeat`](internal/cache/online_status.go:97) - 更新心跳
- [`IsOnline`](internal/cache/online_status.go:62) - 检查用户是否在线
- [`CleanupExpiredUsers`](internal/cache/online_status.go:160) - 清理过期用户（按分数范围删除）

### 缓存使用场景

| 服务/模块 | 使用的缓存管理器 | 文件位置 |
|---------|----------------|---------|
| [`UserService`](internal/service/user_service.go:50) | [`UserCacheManager`](internal/cache/user_cache.go:28), [`FriendCacheManager`](internal/cache/friend_cache.go:27) | [`internal/service/user_service.go:58-59`](internal/service/user_service.go:58) |
| [`MessageService`](internal/service/message_service.go:20) | [`MessageCacheManager`](internal/cache/message_cache.go:39), [`ConversationCacheManager`](internal/cache/conversation_cache.go:35), [`UserCacheManager`](internal/cache/user_cache.go:28) | [`internal/service/message_service.go:29-31`](internal/service/message_service.go:29) |
| [`AuthService`](internal/service/auth_service.go:46) | [`SessionManager`](internal/cache/session.go:34) | [`internal/service/auth_service.go:54`](internal/service/auth_service.go:54) |
| [`WebSocket Manager`](internal/websocket/manager.go:18) | [`OnlineStatusManager`](internal/cache/online_status.go:20) | [`internal/websocket/manager.go:38`](internal/websocket/manager.go:38) |
| [`WebSocket Connection`](internal/websocket/connection.go:31) | [`OnlineStatusManager`](internal/cache/online_status.go:20) | [`internal/websocket/connection.go:50`](internal/websocket/connection.go:50) |

### 设计特点

1. **分层设计**：基础 Redis 封装层 + 业务缓存管理器层
2. **统一键管理**：使用 [`CacheKey`](internal/cache/redis.go:351) 生成器统一管理缓存键
3. **错误处理**：使用 [`CacheError`](internal/cache/redis.go:419) 包装缓存错误
4. **性能优化**：使用 Pipeline 批量操作，减少网络往返
5. **自动过期**：所有缓存都有合理的 TTL 配置
6. **Cache-Aside 模式**：用户缓存支持 [`GetOrLoadUserInfo`](internal/cache/user_cache.go:227) 模式
7. **设备限制**：会话管理器自动清理超过设备数量限制的旧会话

## 开发说明

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 保持函数和变量命名清晰

### 日志

项目使用 Zap 作为日志框架，支持结构化日志记录。

### 数据库

- 使用 GORM 作为 ORM 框架
- 支持数据库迁移和重置
- 支持读写分离

### 中间件

- CORS 跨域处理
- JWT 身份认证
- 日志记录
- 错误恢复
- **接口限流** - 基于 Redis 的滑动窗口限流

## 限流功能

### 限流策略

项目实现了基于 Redis 的两种限流算法：

#### 1. 滑动窗口限流（推荐）

使用 Redis 有序集合（ZSET）+ Lua 脚本实现精确的滑动窗口：

```lua
-- 核心逻辑
1. 删除时间窗口外的记录（ZREMRANGEBYSCORE）
2. 获取当前请求数（ZCARD）
3. 如果未超限：添加当前请求（ZADD）+ 设置过期时间（EXPIRE）
4. 如果超限：返回最早的请求时间作为重置时间
```

**优点**：
- 精确的滑动窗口，不会出现固定窗口的边界问题
- 原子性操作，避免并发问题

#### 2. 简单固定窗口限流

使用 Redis INCR + EXPIRE + Pipeline：

```go
// 核心逻辑
pipe.Incr(ctx, key)        // 递增计数器
pipe.Expire(ctx, key, window)  // 设置过期时间
```

**优点**：性能更高，网络往返少
**缺点**：窗口边界可能出现突发流量

### 限流配置

| 接口类型 | 时间窗口 | 请求限制 | Redis Key 前缀 |
|---------|---------|---------|---------------|
| 认证接口 | 1分钟 | 10次 | `rate_limit:auth:` |
| 消息接口 | 1分钟 | 60次 | `rate_limit:message:` |
| 通用接口 | 1分钟 | 60次 | `rate_limit:general:` |
| 上传接口 | 1分钟 | 10次 | `rate_limit:upload:` |

### 限流键生成策略

按优先级生成限流键：

```go
1. 优先级最高：用户ID → "{prefix}:user:{userID}"
2. 其次：IP地址 → "{prefix}:ip:{ip}"
3. 最后：默认键 → "{prefix}:default"
```

### 响应头

| 响应头 | 说明 |
|--------|------|
| `X-RateLimit-Limit` | 时间窗口内的最大请求次数 |
| `X-RateLimit-Remaining` | 剩余可用请求次数 |
| `X-RateLimit-Reset` | 窗口重置时间戳 |

### 超限响应

返回 HTTP 429 状态码：

```json
{
  "error": "Too many requests",
  "retry_after": 45
}
```

### 容错机制

限流检查失败时，系统会记录错误但允许请求通过，确保 Redis 故障时不会影响业务可用性。

### 代码实现

- 限流中间件：[`internal/middleware/rate_limit.go`](internal/middleware/rate_limit.go)
- 路由配置：[`internal/router/router.go`](internal/router/router.go)

## 健康检查

服务启动后，可以通过以下端点检查服务状态：

```bash
curl http://localhost:8080/health
```

返回示例：

```json
{
  "status": "healthy",
  "services": {
    "postgres": "healthy",
    "redis": "healthy"
  }
}
```

## 许可证

MIT License
