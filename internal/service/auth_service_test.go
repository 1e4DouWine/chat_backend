package service

import (
	"context"
	"testing"

	"chat_backend/internal/dto"
	"chat_backend/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 测试注册功能
func TestAuthService_Register(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return((*model.User)(nil), gorm.ErrRecordNotFound) // 用户不存在
	mockDO.On("Create", mock.AnythingOfType("*model.User")).Return(nil) // 创建成功

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例
	req := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.Register(ctx, req)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试用户名已存在的情况
func TestAuthService_Register_UsernameExists(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望：用户已存在
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{ID: "123", Username: "testuser"}, nil) // 用户已存在

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例
	req := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.Register(ctx, req)
	
	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUsernameAlreadyExists, err)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试登录功能
func TestAuthService_Login(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 加密密码
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// 设置期望
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{
		ID:           "123",
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}, nil)

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例
	req := dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.Login(ctx, req)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试登录时用户名或密码错误的情况
func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 设置期望：用户不存在
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return((*model.User)(nil), gorm.ErrRecordNotFound)

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例
	req := dto.LoginRequest{
		Username: "nonexistent",
		Password: "wrongpassword",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.Login(ctx, req)
	
	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidUsernameOrPassword, err)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试密码验证失败的情况
func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 加密密码
	password := "correctpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// 设置期望：用户存在但密码错误
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{
		ID:           "123",
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}, nil)

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例 - 提供错误密码
	req := dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.Login(ctx, req)
	
	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidUsernameOrPassword, err)

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试刷新令牌功能
func TestAuthService_RefreshToken(t *testing.T) {
	// 创建模拟对象
	mockDB, mockQuery, mockDO := SetupMockDB()
	
	// 创建一个有效的JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "123",
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))

	// 设置期望
	mockQuery.On("User").Return(mockDO)
	mockDO.On("WithContext", mock.Anything).Return(mockDO)
	mockDO.On("Where", mock.Anything).Return(mockDO)
	mockDO.On("First").Return(&model.User{
		ID:       "123",
		Username: "testuser",
	}, nil)

	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例
	req := dto.RefreshRequest{
		RefreshToken: tokenString,
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.RefreshToken(ctx, req)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	// 注意：刷新令牌成功后不会返回新的刷新令牌

	// 验证模拟对象的调用
	mockQuery.AssertExpectations(t)
	mockDO.AssertExpectations(t)
}

// 测试无效的刷新令牌
func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	// 创建服务实例
	authService := &AuthService{db: nil}

	// 创建测试用例 - 无效的令牌
	req := dto.RefreshRequest{
		RefreshToken: "invalid-token",
	}

	ctx := context.Background()
	
	// 运行测试
	resp, err := authService.RefreshToken(ctx, req)
	
	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid refresh token")

	// 验证模拟对象的调用
}