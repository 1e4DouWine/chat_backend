# 群组模块API文档

## 认证说明
所有接口均需要在请求头中携带JWT令牌：
```
Authorization: Bearer {access_token}
```

## 群组接口

### 创建群组
- **接口地址**: `/api/v1/group`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "name": "string" // 群组名称，1-50字符
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "string",
    "name": "string",
    "owner_id": "string",
    "member_count": 1,
    "created_at": "2006-01-02T15:04:05Z07:00"
  }
}
```

### 获取群组列表
- **接口地址**: `/api/v1/group?role={role}`
- **请求方式**: `GET`
- **请求参数**: `role` (查询参数，可选)
  - `owner`: 仅返回自己创建的群组
  - `member`: 仅返回自己加入的群组
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "group_id": "string",
      "name": "string",
      "owner_id": "string",
      "member_count": 5,
      "role": "owner",
      "created_at": "2006-01-02T15:04:05Z07:00"
    }
  ]
}
```

### 获取群组详情
- **接口地址**: `/api/v1/group/{id}`
- **请求方式**: `GET`
- **路径参数**: `id` (群组ID)
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "string",
    "name": "string",
    "owner_id": "string",
    "owner_name": "string",
    "member_count": 5,
    "created_at": "2006-01-02T15:04:05Z07:00",
    "members": [
      {
        "user_id": "string",
        "username": "string",
        "role": "owner",
        "joined_at": "2006-01-02T15:04:05Z07:00"
      }
    ]
  }
}
```

### 加入群组（未启用）
- **接口地址**: `/api/v1/group/{id}/join`
- **请求方式**: `POST`
- **路径参数**: `id` (群组ID)
- **请求参数**:
```json
{
  "invite_code": "string" // 邀请码（可选）
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "string",
    "name": "string",
    "status": "joined"
  }
}
```

### 退出群组
- **接口地址**: `/api/v1/group/{id}/leave`
- **请求方式**: `POST`
- **路径参数**: `id` (群组ID)
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 解散群组
- **接口地址**: `/api/v1/group/{id}`
- **请求方式**: `DELETE`
- **路径参数**: `id` (群组ID)
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 转让群组
- **接口地址**: `/api/v1/group/{id}/transfer`
- **请求方式**: `POST`
- **路径参数**: `id` (群组ID)
- **请求参数**:
```json
{
  "new_owner_id": "string" // 新群主用户ID
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 移除群组成员
- **接口地址**: `/api/v1/group/{group_id}/member/{user_id}`
- **请求方式**: `DELETE`
- **路径参数**: 
  - `group_id`: 群组ID
  - `user_id`: 目标用户ID
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 通过邀请码加入群组
- **接口地址**: `/api/v1/group/join-by-code`
- **请求方式**: `POST`
- **请求参数**:
```json
{
  "invite_code": "string" // 邀请码
}
```
- **响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "string",
    "name": "string",
    "status": "joined"
  }
}
```

## 错误码说明

| 错误码 | 描述 |
|--------|------|
| 40001  | 参数错误 |
| 40002  | 缺少必填字段 |
| 40003  | 群组名称为空 |
| 40301  | 权限不足 |
| 40401  | 群组不存在 |
| 40901  | 已在群组中 |
| 50001  | 创建群组失败 |
| 50002  | 加入群组失败 |
| 50003  | 退出群组失败 |
| 50004  | 解散群组失败 |
| 50005  | 转让群组失败 |
| 50006  | 移除成员失败 |