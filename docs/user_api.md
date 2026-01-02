# 用户模块API文档

## 认证接口

### 注册接口
- **接口地址**: `/api/v1/auth/register`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "username": "string", // 用户名，长度3-20位
  "password": "string"  // 密码，长度6-20位
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### 登录接口
- **接口地址**: `/api/v1/auth/login`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "username": "string", // 用户名
  "password": "string"  // 密码
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### 刷新令牌接口
- **接口地址**: `/api/v1/auth/refresh`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "refresh_token": "string" // 刷新令牌
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

## 用户接口

> **认证说明**: 所有用户接口均需要在请求头中携带JWT令牌
> ```
> Authorization: Bearer {access_token}
> ```

### 获取当前用户信息
- **接口地址**: `/api/v1/user/me`
- **请求方式**: `GET`
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "张三",
    "avatar": "https://ui-avatars.com/api/?name=张三&background=3b82f6&rounded=true&size=128"
  }
}
```

### 搜索用户
- **接口地址**: `/api/v1/user/search?username={username}`
- **请求方式**: `GET`
- **查询参数**:
  - `username`: 用户名（必填）
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "username": "李四",
    "avatar": "https://ui-avatars.com/api/?name=李四&background=10b981&rounded=true&size=128"
  }
}
```

### 获取好友列表
- **接口地址**: `/api/v1/user/friend?status={status}`
- **请求方式**: `GET`
- **查询参数**:
  - `status`: 好友状态（可选，默认为normal）
    - `normal`: 已接受的好友
    - `pending`: 待处理的好友申请
- **响应示例** (status=normal):
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "李四",
      "avatar": "https://ui-avatars.com/api/?name=李四&background=10b981&rounded=true&size=128",
      "status": "normal",
      "create_at": 1234567890
    }
  ]
}
```
- **响应示例** (status=pending):
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440002",
      "username": "王五",
      "avatar": "https://ui-avatars.com/api/?name=王五&background=f59e0b&rounded=true&size=128",
      "status": "pending",
      "create_at": 1234567890
    }
  ]
}
```

### 添加好友
- **接口地址**: `/api/v1/user/friend`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "username": "string", // 目标用户名（必填）
  "message": "string"   // 好友申请消息（选填）
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "pending"
  }
}
```

### 处理好友申请
- **接口地址**: `/api/v1/user/friend/{id}?action={action}`
- **请求方式**: `PUT`
- **路径参数**:
  - `id`: 申请者用户ID
- **查询参数**:
  - `action`: 操作类型（必填）
    - `accept`: 接受好友申请
    - `reject`: 拒绝好友申请
- **响应示例** (接受申请):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_id": "550e8400-e29b-41d4-a716-446655440002",
    "status": "normal"
  }
}
```
- **响应示例** (拒绝申请):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_id": "550e8400-e29b-41d4-a716-446655440002",
    "status": "rejected"
  }
}
```

### 删除好友
- **接口地址**: `/api/v1/user/friend/{id}`
- **请求方式**: `DELETE`
- **路径参数**:
  - `id`: 好友用户ID
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 好友状态说明

| 状态 | 说明 |
|------|------|
| `normal` | 正常好友关系 |
| `pending` | 待处理的好友申请 |
| `rejected` | 已拒绝的好友申请 |
| `removed` | 已删除的好友关系 |

## 业务规则

1. **好友申请有效期**: 7天，过期后自动失效
2. **不能搜索自己**: 搜索用户时不能搜索自己的用户名
3. **重复申请限制**:
   - 已发送未过期的pending申请，不能再次发送
   - 已被拒绝的申请，需等待7天后才能再次发送
   - 对方已发送未过期的pending申请，不能再次发送（应直接处理）
4. **好友关系验证**:
   - 添加好友前会检查是否已经是好友关系
   - 只能删除状态为`normal`的好友关系
5. **申请处理**:
   - 接受申请后，会删除FriendRequest表中的记录，并在Friend表中创建normal记录
   - 拒绝申请后，会更新FriendRequest记录的状态为rejected

