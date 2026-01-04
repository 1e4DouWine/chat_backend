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

	tokenTypeBearer = "Bearer"
)

const (
	errFailedToHashPassword      = "failed to hash password"
	errFailedToCreateUser        = "failed to create user"
	errFailedToGenerateToken     = "failed to generate access token"
	errFailedToGenerateRefresh   = "failed to generate refresh token"
	errJWTNotInitialized         = "JWT not initialized"
	errUnexpectedSigningMethod   = "unexpected signing method"
	errInvalidRefreshToken       = "invalid refresh token"
	errUserNotFound              = "user not found"
	errInvalidUsernameOrPassword = "invalid username or password"
	errUsernameAlreadyExists     = "username already exists"
)

var (
	ErrUsernameAlreadyExists     = errors.New(errUsernameAlreadyExists)
	ErrInvalidUsernameOrPassword = errors.New(errInvalidUsernameOrPassword)
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
		return nil, fmt.Errorf("%s: %w", errFailedToHashPassword, err)
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err = do.Create(&user); err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToCreateUser, err)
	}

	// 生成令牌
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToGenerateToken, err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToGenerateRefresh, err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tokenExpireTime.Seconds()),
		TokenType:    tokenTypeBearer,
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
		return nil, fmt.Errorf("%s: %w", errFailedToGenerateToken, err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToGenerateRefresh, err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tokenExpireTime.Seconds()),
		TokenType:    tokenTypeBearer,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	// 验证刷新令牌
	jwtConfig := middleware.GetJWTConfig()
	if jwtConfig == nil {
		return nil, errors.New(errJWTNotInitialized)
	}

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(errUnexpectedSigningMethod)
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New(errInvalidRefreshToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errInvalidRefreshToken)
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New(errInvalidRefreshToken)
	}

	// 查找用户
	q := dao.Use(s.db).User
	do := q.WithContext(ctx)
	user, err := do.Where(q.ID.Eq(userID)).First()
	if err != nil {
		return nil, errors.New(errUserNotFound)
	}

	// 生成新的访问令牌
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToGenerateToken, err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(tokenExpireTime.Seconds()),
		TokenType:   tokenTypeBearer,
	}, nil
}
