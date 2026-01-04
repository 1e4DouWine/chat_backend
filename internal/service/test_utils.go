package service

import (
	"context"
	"testing"

	"chat_backend/internal/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserModel 是 User 模型的模拟
type MockUserModel struct {
	mock.Mock
}

// MockFriendModel 是 Friend 模型的模拟
type MockFriendModel struct {
	mock.Mock
}

// MockFriendRequestModel 是 FriendRequest 模型的模拟
type MockFriendRequestModel struct {
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

// 通用测试工具函数
func SetupMockDB() (*MockDB, *MockQuery, *MockDAO) {
	mockDB := new(MockDB)
	mockQuery := new(MockQuery)
	mockDO := new(MockDAO)
	
	return mockDB, mockQuery, mockDO
}