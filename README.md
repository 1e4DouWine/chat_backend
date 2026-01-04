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

- 数据持久化
  - PostgreSQL 数据库
  - Redis 缓存（暂未实现）
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

### 认证相关

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新 Token

### 用户相关

- `GET /api/v1/user/me` - 获取当前用户信息
- `GET /api/v1/user/search` - 搜索用户
- `GET /api/v1/user/friend` - 获取好友列表
- `POST /api/v1/user/friend` - 添加好友
- `PUT /api/v1/user/friend/:id` - 处理好友请求
- `DELETE /api/v1/user/friend/:id` - 删除好友

### 群组相关

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

### 消息相关

- `GET /api/v1/message/conversations` - 获取会话列表
- `GET /api/v1/message/private` - 获取私聊消息记录
- `GET /api/v1/message/group/:id` - 获取群聊消息记录

### WebSocket

- `GET /ws` - WebSocket 连接（需要 JWT 认证）
- `GET /api/v1/ws/online` - 获取在线用户列表
- `GET /api/v1/ws/online/:id` - 查询用户是否在线

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
