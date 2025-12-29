# WebSocket 实时通信 API 文档

## 概述

本文档描述了聊天系统WebSocket实时通信接口的规范。WebSocket用于实现客户端与服务端之间的双向实时消息传输。

## 连接信息

### 连接地址

```
ws://localhost:8080/ws
```

生产环境使用 `wss://` 协议。

### 连接方式

WebSocket连接需要通过HTTP升级请求建立，连接时需要在URL参数中携带JWT token进行身份验证。

```
ws://localhost:8080/ws?token={access_token}
```

**参数说明：**
- `token`: JWT访问令牌，通过登录接口获取

### 连接示例

```javascript
// JavaScript 客户端连接示例
const token = localStorage.getItem('access_token');
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = function() {
    console.log('WebSocket连接已建立');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    handleMessage(message);
};

ws.onerror = function(error) {
    console.error('WebSocket错误:', error);
};

ws.onclose = function(event) {
    console.log('WebSocket连接已关闭:', event.code, event.reason);
};
```

---

## 消息格式

### 通用消息结构

所有WebSocket消息均采用JSON格式，包含以下字段：

```json
{
    "type": "message_type",
    "data": {},
    "timestamp": 1234567890,
    "message_id": "uuid"
}
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 消息类型，见下方消息类型定义 |
| data | object | 是 | 消息数据，根据type不同而不同 |
| timestamp | int64 | 是 | Unix时间戳（秒） |
| message_id | string | 是 | 消息唯一标识（UUID格式） |

---

## 消息类型定义

### 1. 心跳消息

#### 1.1 客户端发送心跳（Ping）

**消息类型**: `ping`

**请求示例：**
```json
{
    "type": "ping",
    "data": {},
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**服务端响应（Pong）：**
```json
{
    "type": "pong",
    "data": {
        "server_time": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**说明：**
- 客户端应每隔30-60秒发送一次ping消息
- 服务端收到ping后立即返回pong
- 如果60秒内未收到pong，客户端应尝试重连

---

### 2. 私聊消息

#### 2.1 发送私聊消息

**消息类型**: `private_message`

**请求示例：**
```json
{
    "type": "private_message",
    "data": {
        "to_user_id": "550e8400-e29b-41d4-a716-446655440000",
        "content": "你好，这是一条私聊消息",
        "message_type": "text"
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440002"
}
```

**data字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| to_user_id | string | 是 | 接收者用户ID |
| content | string | 是 | 消息内容 |
| message_type | string | 是 | 消息类型：text/image/file等 |

**服务端响应（消息发送成功）：**
```json
{
    "type": "message_sent",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440002",
        "status": "sent",
        "sent_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

**接收者收到的消息：**
```json
{
    "type": "private_message",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440002",
        "from_user_id": "550e8400-e29b-41d4-a716-446655440000",
        "from_username": "张三",
        "from_avatar": "https://ui-avatars.com/api/?name=张三&background=...",
        "content": "你好，这是一条私聊消息",
        "message_type": "text",
        "created_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440004"
}
```

---

### 3. 群聊消息

#### 3.1 发送群聊消息

**消息类型**: `group_message`

**请求示例：**
```json
{
    "type": "group_message",
    "data": {
        "group_id": "660e8400-e29b-41d4-a716-446655440000",
        "content": "大家好，这是一条群聊消息",
        "message_type": "text"
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440005"
}
```

**data字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群组ID |
| content | string | 是 | 消息内容 |
| message_type | string | 是 | 消息类型：text/image/file等 |

**服务端响应（消息发送成功）：**
```json
{
    "type": "message_sent",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440005",
        "status": "sent",
        "sent_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440006"
}
```

**群组成员收到的消息：**
```json
{
    "type": "group_message",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440005",
        "group_id": "660e8400-e29b-41d4-a716-446655440000",
        "group_name": "技术交流群",
        "from_user_id": "550e8400-e29b-41d4-a716-446655440000",
        "from_username": "张三",
        "from_avatar": "https://ui-avatars.com/api/?name=张三&background=...",
        "content": "大家好，这是一条群聊消息",
        "message_type": "text",
        "created_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440007"
}
```

---

### 4. 消息已读确认

#### 4.1 标记消息为已读

**消息类型**: `message_read`

**请求示例：**
```json
{
    "type": "message_read",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440002",
        "chat_type": "private",
        "target_id": "550e8400-e29b-41d4-a716-446655440000"
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440008"
}
```

**data字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | string | 是 | 消息ID |
| chat_type | string | 是 | 聊天类型：private（私聊）或group（群聊） |
| target_id | string | 是 | 目标ID（私聊为用户ID，群聊为群组ID） |

**服务端响应：**
```json
{
    "type": "read_receipt",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440002",
        "read_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440009"
}
```

**消息发送者收到的已读通知：**
```json
{
    "type": "message_read",
    "data": {
        "message_id": "550e8400-e29b-41d4-a716-446655440002",
        "read_by": "660e8400-e29b-41d4-a716-446655440000",
        "read_by_username": "李四",
        "read_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440010"
}
```

---

### 5. 系统消息

#### 5.1 好友申请通知

**消息类型**: `system_message`

**通知示例：**
```json
{
    "type": "system_message",
    "data": {
        "system_type": "friend_request",
        "from_user_id": "550e8400-e29b-41d4-a716-446655440000",
        "from_username": "张三",
        "from_avatar": "https://ui-avatars.com/api/?name=张三&background=...",
        "message": "请求添加您为好友",
        "created_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440013"
}
```

#### 5.2 好友申请已接受通知

```json
{
    "type": "system_message",
    "data": {
        "system_type": "friend_accepted",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "张三",
        "avatar": "https://ui-avatars.com/api/?name=张三&background=...",
        "message": "已接受您的好友申请",
        "created_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440014"
}
```

#### 5.3 群组邀请通知

```json
{
    "type": "system_message",
    "data": {
        "system_type": "group_invite",
        "group_id": "660e8400-e29b-41d4-a716-446655440000",
        "group_name": "技术交流群",
        "from_user_id": "550e8400-e29b-41d4-a716-446655440000",
        "from_username": "张三",
        "message": "邀请您加入群组",
        "created_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440015"
}
```

---

### 6. 在线状态

#### 6.1 用户上线通知

**消息类型**: `user_online`

**通知示例：**
```json
{
    "type": "user_online",
    "data": {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "张三",
        "online_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440017"
}
```

#### 6.2 用户下线通知

**消息类型**: `user_offline`

**通知示例：**
```json
{
    "type": "user_offline",
    "data": {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "张三",
        "offline_at": 1234567890
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440018"
}
```

---

### 7. 离线消息推送

#### 7.1 离线消息列表

**消息类型**: `offline_messages`

**推送示例：**
```json
{
    "type": "offline_messages",
    "data": {
        "total": 5,
        "messages": [
            {
                "message_id": "550e8400-e29b-41d4-a716-446655440020",
                "chat_type": "private",
                "from_user_id": "550e8400-e29b-41d4-a716-446655440000",
                "from_username": "张三",
                "from_avatar": "https://ui-avatars.com/api/?name=张三&background=...",
                "content": "你好，这是一条离线消息",
                "message_type": "text",
                "created_at": 1234567890
            },
            {
                "message_id": "550e8400-e29b-41d4-a716-446655440021",
                "chat_type": "group",
                "group_id": "660e8400-e29b-41d4-a716-446655440000",
                "group_name": "技术交流群",
                "from_user_id": "550e8400-e29b-41d4-a716-446655440001",
                "from_username": "李四",
                "from_avatar": "https://ui-avatars.com/api/?name=李四&background=...",
                "content": "群聊离线消息",
                "message_type": "text",
                "created_at": 1234567891
            }
        ]
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440019"
}
```

**说明：**
- 用户上线后，服务端会推送离线期间收到的所有消息
- 离线消息按时间顺序排列
- 推送完成后，服务端会清除已推送的离线消息

---

## 错误处理

### 错误消息格式

```json
{
    "type": "error",
    "data": {
        "code": 40001,
        "message": "参数错误",
        "details": "缺少必填字段: to_user_id"
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440099"
}
```

### 错误码定义

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数错误 |
| 40002 | 缺少必填字段 |
| 40003 | 消息格式错误 |
| 40101 | 未授权（token无效或过期） |
| 40102 | 认证失败 |
| 40301 | 权限不足 |
| 40401 | 用户不存在 |
| 40402 | 群组不存在 |
| 40901 | 消息ID重复 |
| 42901 | 发送频率过高 |
| 50001 | 服务器内部错误 |
| 50002 | 消息发送失败 |
| 50003 | 数据库操作失败 |

### WebSocket关闭码

| 关闭码 | 说明 |
|--------|------|
| 1000 | 正常关闭 |
| 1001 | 端点离开 |
| 1002 | 协议错误 |
| 1003 | 不支持的数据类型 |
| 1008 | 策略违规 |
| 1011 | 服务器错误 |
| 4000 | 未授权 |
| 4001 | Token过期 |
| 4002 | 频率限制 |
| 4003 | 连接被服务器关闭 |

---

## 连接管理

### 连接建立流程

1. 客户端发起WebSocket连接请求，携带JWT token
2. 服务端验证token有效性
3. 验证成功后建立连接，返回连接成功消息
4. 客户端开始发送心跳消息
5. 双方开始正常通信

### 连接成功响应

```json
{
    "type": "connected",
    "data": {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "张三",
        "server_time": 1234567890,
        "heartbeat_interval": 30
    },
    "timestamp": 1234567890,
    "message_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 连接断开处理

**正常断开：**
- 客户端主动关闭连接（发送关闭帧）
- 服务端清理连接资源
- 通知好友用户下线

**异常断开：**
- 网络中断或超时
- 服务端检测到心跳超时
- 服务端主动关闭连接
- 客户端应尝试重新连接

### 重连策略

1. 指数退避重连：1s, 2s, 4s, 8s, 16s, 32s
2. 最大重试次数：5次
3. 重连成功后，服务端推送离线消息

---

## 消息类型汇总

| 消息类型 | 方向 | 说明 |
|----------|------|------|
| connected | 服务端→客户端 | 连接成功 |
| ping | 客户端→服务端 | 心跳请求 |
| pong | 服务端→客户端 | 心跳响应 |
| private_message | 双向 | 私聊消息 |
| group_message | 双向 | 群聊消息 |
| message_sent | 服务端→客户端 | 消息发送确认 |
| message_read | 双向 | 消息已读确认 |
| read_receipt | 服务端→客户端 | 已读回执 |
| typing | 双向 | 输入状态 |
| system_message | 服务端→客户端 | 系统通知 |
| user_online | 服务端→客户端 | 用户上线 |
| user_offline | 服务端→客户端 | 用户下线 |
| offline_messages | 服务端→客户端 | 离线消息推送 |
| error | 服务端→客户端 | 错误消息 |

---

## 消息类型定义

### message_type 枚举

| 值 | 说明 |
|----|------|
| text | 文本消息 |
| image | 图片消息 |
| file | 文件消息 |
| voice | 语音消息 |
| video | 视频消息 |
| location | 位置消息 |

### system_type 枚举

| 值 | 说明 |
|----|------|
| friend_request | 好友申请 |
| friend_accepted | 好友申请已接受 |
| friend_rejected | 好友申请已拒绝 |
| group_invite | 群组邀请 |
| group_joined | 加入群组 |
| group_left | 退出群组 |
| group_disbanded | 群组解散 |
| message_retracted | 消息撤回 |

---

## 完整示例

### JavaScript 客户端完整示例

```javascript
class ChatWebSocket {
    constructor(token) {
        this.token = token;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.heartbeatInterval = null;
        this.messageHandlers = {};
    }

    connect() {
        const wsUrl = `ws://localhost:8080/ws?token=${this.token}`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket连接已建立');
            this.reconnectAttempts = 0;
            this.startHeartbeat();
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket错误:', error);
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket连接已关闭:', event.code, event.reason);
            this.stopHeartbeat();
            this.reconnect();
        };
    }

    handleMessage(message) {
        const handler = this.messageHandlers[message.type];
        if (handler) {
            handler(message.data);
        }

        switch (message.type) {
            case 'connected':
                console.log('连接成功:', message.data);
                break;
            case 'private_message':
                this.onPrivateMessage(message.data);
                break;
            case 'group_message':
                this.onGroupMessage(message.data);
                break;
            case 'system_message':
                this.onSystemMessage(message.data);
                break;
            case 'error':
                this.onError(message.data);
                break;
        }
    }

    sendPrivateMessage(toUserId, content) {
        const message = {
            type: 'private_message',
            data: {
                to_user_id: toUserId,
                content: content,
                message_type: 'text'
            },
            timestamp: Math.floor(Date.now() / 1000),
            message_id: this.generateUUID()
        };
        this.ws.send(JSON.stringify(message));
    }

    sendGroupMessage(groupId, content) {
        const message = {
            type: 'group_message',
            data: {
                group_id: groupId,
                content: content,
                message_type: 'text'
            },
            timestamp: Math.floor(Date.now() / 1000),
            message_id: this.generateUUID()
        };
        this.ws.send(JSON.stringify(message));
    }

    sendTyping(chatType, targetId, isTyping) {
        const message = {
            type: 'typing',
            data: {
                chat_type: chatType,
                target_id: targetId,
                is_typing: isTyping
            },
            timestamp: Math.floor(Date.now() / 1000),
            message_id: this.generateUUID()
        };
        this.ws.send(JSON.stringify(message));
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            const ping = {
                type: 'ping',
                data: {},
                timestamp: Math.floor(Date.now() / 1000),
                message_id: this.generateUUID()
            };
            this.ws.send(JSON.stringify(ping));
        }, 30000);
    }

    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }

    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            const delay = Math.pow(2, this.reconnectAttempts) * 1000;
            setTimeout(() => {
                this.reconnectAttempts++;
                console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
                this.connect();
            }, delay);
        } else {
            console.error('达到最大重连次数，停止重连');
        }
    }

    close() {
        this.stopHeartbeat();
        if (this.ws) {
            this.ws.close();
        }
    }

    generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    onPrivateMessage(data) {
        console.log('收到私聊消息:', data);
    }

    onGroupMessage(data) {
        console.log('收到群聊消息:', data);
    }

    onSystemMessage(data) {
        console.log('收到系统消息:', data);
    }

    onError(data) {
        console.error('收到错误消息:', data);
    }
}

// 使用示例
const token = localStorage.getItem('access_token');
const chat = new ChatWebSocket(token);
chat.connect();

// 发送私聊消息
chat.sendPrivateMessage('user-id-123', '你好');

// 发送群聊消息
chat.sendGroupMessage('group-id-456', '大家好');

// 发送输入状态
chat.sendTyping('private', 'user-id-123', true);
setTimeout(() => {
    chat.sendTyping('private', 'user-id-123', false);
}, 3000);
```

---

## 注意事项

1. **Token管理**: JWT token过期后需要通过HTTP API刷新，然后重新建立WebSocket连接
2. **消息ID**: 每条消息必须有唯一的message_id，用于消息去重和追踪
3. **时间戳**: 所有消息必须包含timestamp字段，使用Unix时间戳（秒）
4. **心跳保活**: 客户端必须定期发送心跳消息，否则连接会被服务端关闭
5. **重连机制**: 网络不稳定时，客户端应实现自动重连机制
6. **消息顺序**: 服务端保证消息按时间顺序推送，客户端应按顺序处理
7. **错误处理**: 客户端应妥善处理各种错误情况，提供友好的用户提示
8. **安全性**: 生产环境必须使用wss协议，确保通信安全

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.0 | 2025-12-29 | 初始版本 |
