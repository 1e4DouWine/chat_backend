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

### 获取当前用户信息
- **接口地址**: `/api/v1/user/me`
- **请求方式**: `GET`
- **请求头**: `Authorization: Bearer {access_token}`
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "string",
    "username": "string",
    "avatar": "string"
  }
}
```

### 搜索用户
- **接口地址**: `/api/v1/user/search?username={username}`
- **请求方式**: `GET`
- **请求头**: `Authorization: Bearer {access_token}`
- **请求参数**: `username` (查询参数，目标用户名)
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "string",
    "username": "string",
    "avatar": "string"
  }
}
```

### 添加好友
- **接口地址**: `/api/v1/user/friend`
- **请求方式**: `POST`
- **请求头**: `Authorization: Bearer {access_token}`
- **请求参数**:
```json
{
  "username": "string", // 目标用户名
  "message": "string"   // 好友申请消息（可选）
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_id": "string",
    "status": "pending"
  }
}
```

### 获取好友列表
- **接口地址**: `/api/v1/user/friend?status={status}`
- **请求方式**: `GET`
- **请求头**: `Authorization: Bearer {access_token}`
- **请求参数**: `status` (查询参数，可选，默认值：accepted)
  - `pending`: 待处理的好友申请
  - `accepted`: 已通过的好友
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "user_id": "string",
      "username": "string",
      "avatar": "string",
      "status": "string",
      "create_at": 1234567890
    }
  ]
}
```

### 处理好友申请
- **接口地址**: `/api/v1/user/friend/{id}?action={action}`
- **请求方式**: `PUT`
- **请求头**: `Authorization: Bearer {access_token}`
- **路径参数**: `id` (好友ID)
- **查询参数**: `action`
  - `accept`: 接受好友申请
  - `reject`: 拒绝好友申请
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_id": "string",
    "status": "accepted"// rejected
  }
}
```

### 删除好友
- **接口地址**: `/api/v1/user/friend/{id}`
- **请求方式**: `DELETE`
- **请求头**: `Authorization: Bearer {access_token}`
- **路径参数**: `id` (好友ID)
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 错误码说明

| 错误码 | 描述 |
|--------|------|
| 40001  | 参数错误 |
| 40002  | 缺少必填字段 |
| 40101  | 用户名或密码错误 |
| 40102  | 用户名已存在 |
| 40103  | 无效的刷新令牌 |
| 40301  | 用户ID不能为空 |
| 40401  | 用户不存在 |
| 50001  | 注册失败 |
| 50002  | 登录失败 |
| 50003  | 获取用户信息失败 |
| 50004  | 添加好友失败 |
| 50005  | 获取好友列表失败 |
| 50006  | 处理好友申请失败 |
| 50007  | 删除好友失败 |