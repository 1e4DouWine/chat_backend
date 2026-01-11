package service

import (
	"chat_backend/internal/cache"
	"chat_backend/internal/config"
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
	db             *gorm.DB
	sessionManager *cache.SessionManager
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:             db,
		sessionManager: cache.NewSessionManager(),
	}
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

	// 将 Refresh Token 存储到 Redis
	deviceInfo := map[string]string{
		"device_id":   req.DeviceID,
		"device_name": req.DeviceName,
		"user_agent":  req.UserAgent,
		"ip_address":  req.IPAddress,
	}
	cfg := config.GetConfig()
	if err := s.sessionManager.StoreRefreshToken(ctx, user.ID, user.Username, refreshToken, time.Duration(cfg.JWT.RefreshExpiry)*time.Hour, deviceInfo); err != nil {
		// Redis 存储失败不影响注册流程，记录日志即可
		fmt.Printf("Warning: failed to store refresh token in redis: %v\n", err)
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

	// 将 Refresh Token 存储到 Redis
	deviceInfo := map[string]string{
		"device_id":   req.DeviceID,
		"device_name": req.DeviceName,
		"user_agent":  req.UserAgent,
		"ip_address":  req.IPAddress,
	}
	cfg := config.GetConfig()
	if err := s.sessionManager.StoreRefreshToken(ctx, user.ID, user.Username, refreshToken, time.Duration(cfg.JWT.RefreshExpiry)*time.Hour, deviceInfo); err != nil {
		// Redis 存储失败不影响登录流程，记录日志即可
		fmt.Printf("Warning: failed to store refresh token in redis: %v\n", err)
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

	// 验证 Refresh Token 是否在 Redis 中存在
	_, err = s.sessionManager.ValidateRefreshToken(ctx, userID, req.RefreshToken)
	if err != nil {
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

	// 可选：生成新的 Refresh Token 并更新 Redis
	// newRefreshToken, err := middleware.GenerateRefreshToken(user.ID)
	// if err != nil {
	// 	return nil, fmt.Errorf("%s: %w", errFailedToGenerateRefresh, err)
	// }
	// if err := s.sessionManager.RefreshSession(ctx, userID, req.RefreshToken, newRefreshToken, time.Duration(cfg.JWT.RefreshExpiry)*time.Hour); err != nil {
	// 	fmt.Printf("Warning: failed to refresh session in redis: %v\n", err)
	// }

	return &dto.AuthResponse{
		AccessToken: accessToken,
		// RefreshToken: newRefreshToken, // 如果需要更新 Refresh Token，取消注释
		ExpiresIn: int64(tokenExpireTime.Seconds()),
		TokenType: tokenTypeBearer,
	}, nil
}

// RevokeToken 撤销 Refresh Token
func (s *AuthService) RevokeToken(ctx context.Context, userID, refreshToken string) error {
	return s.sessionManager.RevokeRefreshToken(ctx, userID, refreshToken)
}

// RevokeAllUserTokens 撤销用户的所有会话
func (s *AuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.sessionManager.RevokeAllUserSessions(ctx, userID)
}

// GetUserSessions 获取用户的所有活跃会话
func (s *AuthService) GetUserSessions(ctx context.Context, userID string) ([]cache.SessionInfo, error) {
	return s.sessionManager.GetUserSessions(ctx, userID)
}
