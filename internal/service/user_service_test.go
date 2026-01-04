package service

import (
	"context"
	"testing"

	"chat_backend/internal/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 测试获取用户信息功能
func TestUserService_GetMe(t *testing.T) {
	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	username := "testuser"
	
	// 运行测试
	resp, err := userService.GetMe(ctx, userID, username)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, username, resp.Username)
	assert.Contains(t, resp.Avatar, "ui-avatars.com")
}

// 测试获取用户资料功能
func TestUserService_GetUserInfo(t *testing.T) {
	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	username := "testuser"
	
	// 运行测试
	resp, err := userService.GetUserInfo(ctx, userID, username)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, username, resp.Username)
	assert.Contains(t, resp.Avatar, "ui-avatars.com")
}

// 测试通过用户名获取用户ID功能
func TestUserService_GetUserIDByUsername(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{ID: "123", Username: "testuser"}, nil)

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	username := "testuser"
	
	// 运行测试
	result, err := userService.GetUserIDByUsername(ctx, username)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, "123", result)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试用户不存在的情况
func TestUserService_GetUserIDByUsername_UserNotFound(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望：用户不存在
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return((*model.User)(nil), gorm.ErrRecordNotFound)

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	username := "nonexistent"
	
	// 运行测试
	result, err := userService.GetUserIDByUsername(ctx, username)
	
	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试通过用户ID获取用户名功能
func TestUserService_GetUsernameByUserID(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{ID: "123", Username: "testuser"}, nil)

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	
	// 运行测试
	result, err := userService.GetUsernameByUserID(ctx, userID)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, "testuser", result)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试发送好友请求 - 成功情况
func TestUserService_SendAddFriendRequest_Success(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockFriendDO := new(MockDAO)
	mockRequestDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendDO)
	mockFriendDO.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Or", mock.Anything, mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("First").Return((*model.Friend)(nil), gorm.ErrRecordNotFound) // 没有好友关系
	
	mockQuery.On("FriendRequest").Return(mockRequestDO)
	mockRequestDO.On("WithContext", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("First").Return(&model.FriendRequest{Status: "expired"}, gorm.ErrRecordNotFound) // 没有未过期的请求
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("First").Return((*model.FriendRequest)(nil), gorm.ErrRecordNotFound) // 没有收到的请求
	mockRequestDO.On("Create", mock.AnythingOfType("*model.FriendRequest")).Return(nil) // 创建请求成功

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	friendID := "456"
	
	// 运行测试
	resp, err := userService.SendAddFriendRequest(ctx, userID, friendID)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, friendID, resp.FriendID)
	assert.Equal(t, FriendRequestStatusPending, resp.Status)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
}

// 测试发送好友请求 - 已是好友的情况
func TestUserService_SendAddFriendRequest_AlreadyFriends(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockFriendDO := new(MockDAO)
	
	// 设置期望：已是好友
	mockQuery.On("Friend").Return(mockFriendDO)
	mockFriendDO.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Or", mock.Anything, mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("First").Return(&model.Friend{Status: FriendStatusNormal}, nil)

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	friendID := "456"
	
	// 运行测试
	resp, err := userService.SendAddFriendRequest(ctx, userID, friendID)
	
	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), errAlreadyFriends)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试处理好友请求 - 接受请求
func TestUserService_ProcessFriendRequest_Accept(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockRequestDO := new(MockDAO)
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("FriendRequest").Return(mockRequestDO)
	mockRequestDO.On("WithContext", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("First").Return(&model.FriendRequest{Status: FriendRequestStatusPending}, nil) // 请求存在
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Delete").Return(int64(1), nil) // 删除请求成功
	
	mockQuery.On("Friend").Return(mockFriendDO)
	mockFriendDO.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Create", mock.AnythingOfType("*model.Friend")).Return(nil) // 创建好友关系成功

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	friendID := "456"
	action := "accept"
	
	// 运行测试
	resp, err := userService.ProcessFriendRequest(ctx, userID, friendID, action)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, friendID, resp.FriendID)
	assert.Equal(t, FriendStatusNormal, resp.Status)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试处理好友请求 - 拒绝请求
func TestUserService_ProcessFriendRequest_Reject(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockRequestDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("FriendRequest").Return(mockRequestDO)
	mockRequestDO.On("WithContext", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("First").Return(&model.FriendRequest{Status: FriendRequestStatusPending}, nil) // 请求存在
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Update", mock.Anything, mock.Anything).Return(int64(1), nil) // 更新状态成功

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	friendID := "456"
	action := "reject"
	
	// 运行测试
	resp, err := userService.ProcessFriendRequest(ctx, userID, friendID, action)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, friendID, resp.FriendID)
	assert.Equal(t, FriendRequestStatusRejected, resp.Status)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
}

// 测试删除好友功能
func TestUserService_DeleteFriend(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendDO)
	mockFriendDO.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Or", mock.Anything, mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("First").Return(&model.Friend{UserA: "123", UserB: "456", Status: FriendStatusNormal}, nil) // 好友关系存在
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Update", mock.Anything, mock.Anything).Return(int64(1), nil) // 更新状态成功

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	friendID := "456"
	
	// 运行测试
	err := userService.DeleteFriend(ctx, userID, friendID)
	
	// 验证结果
	assert.NoError(t, err)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试检查好友关系功能
func TestUserService_IsFriend(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendDO)
	mockFriendDO.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Or", mock.Anything, mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Count").Return(int64(1), nil) // 找到1个好友关系

	// 创建服务实例
	userService := &UserService{db: nil}

	// 创建测试用例
	ctx := context.Background()
	userID := "123"
	targetUserID := "456"
	
	// 运行测试
	isFriend, err := userService.IsFriend(ctx, userID, targetUserID)
	
	// 验证结果
	assert.NoError(t, err)
	assert.True(t, isFriend)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}