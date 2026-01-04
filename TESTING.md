# 单元测试说明

## 项目概述
本项目包含为聊天后端API编写的单元测试，主要覆盖以下服务：

1. 认证服务 (auth_service)
2. 用户服务 (user_service)

## 测试内容

### 认证服务测试 (auth_service_test.go)
- 用户注册功能测试
- 用户登录功能测试
- 刷新令牌功能测试
- 错误情况测试（如用户名已存在、密码错误等）

### 用户服务测试 (user_service_test.go)
- 获取用户信息功能测试
- 用户ID/用户名查询功能测试
- 好友请求功能测试
- 好友关系管理功能测试
- 工具函数测试（颜色生成、限制函数等）

## 测试结构

### Mock对象
测试使用了模拟对象来模拟数据库操作，包括：
- MockQuery: 模拟数据库查询
- MockDAO: 模拟数据访问对象
- MockDB: 模拟数据库连接

### 测试用例设计
每个测试用例都遵循以下模式：
1. 创建模拟对象
2. 设置期望行为
3. 运行测试函数
4. 验证结果
5. 验证模拟对象调用

## 运行测试

在正常环境下，可以使用以下命令运行测试：

```bash
# 运行所有服务测试
go test ./internal/service/ -v

# 运行特定测试
go test ./internal/service/ -run TestFunctionName -v

# 运行测试并生成覆盖率报告
go test ./internal/service/ -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 注意事项

当前环境存在Go版本兼容性问题，Go 1.19与项目要求的Go 1.25不完全兼容，这可能导致测试无法正常运行。在实际部署环境中，应使用正确的Go版本。

## 测试覆盖范围

- [x] 认证服务功能
- [x] 用户服务功能
- [x] 好友关系管理
- [x] 工具函数
- [ ] 集成测试（待实现）
- [ ] 边界条件测试（待实现）