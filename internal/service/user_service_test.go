package service

import (
	"context"
	"testing"
	"time"

	"chat_backend/internal/dto"
	"chat_backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockFriendModel 是 Friend 模型的模拟
type MockFriendModel struct {
	mock.Mock
}

// MockFriendRequestModel 是 FriendRequest 模型的模拟
type MockFriendRequestModel struct {
	mock.Mock
}

// MockUserModel 是 User 模型的模拟
type MockUserModel struct {
	mock.Mock
}

// MockDAO 是数据访问对象的模拟
type MockDAO struct {
	mock.Mock
}

// 模拟查询方法
func (m *MockDAO) Where(...interface{}) *MockDAO {
	args := m.Called()
	return args.Get(0).(*MockDAO)
}

func (m *MockDAO) First() (*model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockDAO) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockDAO) Scan(dest interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func (m *MockDAO) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDAO) Update(column interface{}, value interface{}) (int64, error) {
	args := m.Called(column, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDAO) Delete() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDAO) Or(...interface{}) *MockDAO {
	args := m.Called()
	return args.Get(0).(*MockDAO)
}

func (m *MockDAO) Gt(interface{}) interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDAO) Eq(interface{}) interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDAO) WithContext(ctx context.Context) *MockDAO {
	args := m.Called(ctx)
	return args.Get(0).(*MockDAO)
}

// MockQuery 是查询接口的模拟
type MockQuery struct {
	mock.Mock
}

func (m *MockQuery) User() *MockDAO {
	args := m.Called()
	return args.Get(0).(*MockDAO)
}

func (m *MockQuery) Friend() *MockDAO {
	args := m.Called()
	return args.Get(0).(*MockDAO)
}

func (m *MockQuery) FriendRequest() *MockDAO {
	args := m.Called()
	return args.Get(0).(*MockDAO)
}

// MockDB 是数据库的模拟
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Use(*gorm.DB) *MockQuery {
	args := m.Called()
	return args.Get(0).(*MockQuery)
}

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
	mockQuery := new(MockQuery)
	mockUserQuery := new(MockDAO)
	mockDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("User").Return(mockUserQuery)
	mockUserQuery.On("WithContext", mock.Anything).Return(mockDO)
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
	mockUserQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试用户不存在的情况
func TestUserService_GetUserIDByUsername_UserNotFound(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockUserQuery := new(MockDAO)
	mockDO := new(MockDAO)
	
	// 设置期望：用户不存在
	mockQuery.On("User").Return(mockUserQuery)
	mockUserQuery.On("WithContext", mock.Anything).Return(mockDO)
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
	mockUserQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试通过用户ID获取用户名功能
func TestUserService_GetUsernameByUserID(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockUserQuery := new(MockDAO)
	mockDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("User").Return(mockUserQuery)
	mockUserQuery.On("WithContext", mock.Anything).Return(mockDO)
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
	mockUserQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试发送好友请求 - 成功情况
func TestUserService_SendAddFriendRequest_Success(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockFriendQuery := new(MockDAO)
	mockFriendDO := new(MockDAO)
	mockRequestQuery := new(MockDAO)
	mockRequestDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendQuery)
	mockFriendQuery.On("WithContext", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Where", mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("Or", mock.Anything, mock.Anything).Return(mockFriendDO)
	mockFriendDO.On("First").Return((*model.Friend)(nil), gorm.ErrRecordNotFound) // 没有好友关系
	
	mockQuery.On("FriendRequest").Return(mockRequestQuery)
	mockRequestQuery.On("WithContext", mock.Anything).Return(mockRequestDO)
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
	mockFriendQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
	mockRequestQuery.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
}

// 测试发送好友请求 - 已是好友的情况
func TestUserService_SendAddFriendRequest_AlreadyFriends(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockFriendQuery := new(MockDAO)
	mockFriendDO := new(MockDAO)
	
	// 设置期望：已是好友
	mockQuery.On("Friend").Return(mockFriendQuery)
	mockFriendQuery.On("WithContext", mock.Anything).Return(mockFriendDO)
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
	mockFriendQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试处理好友请求 - 接受请求
func TestUserService_ProcessFriendRequest_Accept(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockRequestQuery := new(MockDAO)
	mockRequestDO := new(MockDAO)
	mockFriendQuery := new(MockDAO)
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("FriendRequest").Return(mockRequestQuery)
	mockRequestQuery.On("WithContext", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("First").Return(&model.FriendRequest{Status: FriendRequestStatusPending}, nil) // 请求存在
	mockRequestDO.On("Where", mock.Anything).Return(mockRequestDO)
	mockRequestDO.On("Delete").Return(int64(1), nil) // 删除请求成功
	
	mockQuery.On("Friend").Return(mockFriendQuery)
	mockFriendQuery.On("WithContext", mock.Anything).Return(mockFriendDO)
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
	mockRequestQuery.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
	mockFriendQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试处理好友请求 - 拒绝请求
func TestUserService_ProcessFriendRequest_Reject(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockRequestQuery := new(MockDAO)
	mockRequestDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("FriendRequest").Return(mockRequestQuery)
	mockRequestQuery.On("WithContext", mock.Anything).Return(mockRequestDO)
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
	mockRequestQuery.AssertExpectations(t)
	mockRequestDO.AssertExpectations(t)
}

// 测试删除好友功能
func TestUserService_DeleteFriend(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockFriendQuery := new(MockDAO)
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendQuery)
	mockFriendQuery.On("WithContext", mock.Anything).Return(mockFriendDO)
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
	mockFriendQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试检查好友关系功能
func TestUserService_IsFriend(t *testing.T) {
	// 创建模拟对象
	mockQuery := new(MockQuery)
	mockFriendQuery := new(MockDAO)
	mockFriendDO := new(MockDAO)
	
	// 设置期望
	mockQuery.On("Friend").Return(mockFriendQuery)
	mockFriendQuery.On("WithContext", mock.Anything).Return(mockFriendDO)
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
	mockFriendQuery.AssertExpectations(t)
	mockFriendDO.AssertExpectations(t)
}

// 测试颜色生成函数
func TestColorFromUUID(t *testing.T) {
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	
	color := colorFromUUID(uuid)
	
	// 验证颜色是6位十六进制字符串
	assert.Len(t, color, 6)
	
	// 验证颜色在合理范围内（50-200之间的值转换为十六进制）
	for _, c := range color {
		assert.Contains(t, "0123456789abcdef", string(c))
	}
}

// 测试限制函数
func TestClamp(t *testing.T) {
	// 测试值在范围内
	assert.Equal(t, byte(100), clamp(100, 50, 200))
	
	// 测试值小于最小值
	assert.Equal(t, byte(50), clamp(30, 50, 200))
	
	// 测试值大于最大值
	assert.Equal(t, byte(200), clamp(250, 50, 200))
}