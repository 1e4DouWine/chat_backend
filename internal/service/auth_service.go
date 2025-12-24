package service

import (
	"chat_backend/internal/dao"
	"chat_backend/internal/dto"
	"chat_backend/internal/middleware"
	"chat_backend/internal/model"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	tokenExpireTime = time.Hour * 24
)

var (
	ErrUsernameAlreadyExists     = errors.New("username already exists")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)
	// 检查用户名是否已存在
	_, err := do.Where(q.Username.Eq(req.Username)).First()
	if err == nil {
		return nil, ErrUsernameAlreadyExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err = do.Create(&user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 生成令牌
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tokenExpireTime.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)
	// 查找用户
	user, err := do.Where(q.Username.Eq(req.Username)).First()
	if err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	// 验证密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	// 生成令牌
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tokenExpireTime.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	// 验证刷新令牌
	jwtConfig := middleware.GetJWTConfig()
	if jwtConfig == nil {
		return nil, errors.New("JWT not initialized")
	}

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid refresh token")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid refresh token")
	}

	// 查找用户
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)
	user, err := do.Where(q.ID.Eq(userID)).First()
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 生成新的访问令牌
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(tokenExpireTime.Seconds()),
		TokenType:   "Bearer",
	}, nil
}
