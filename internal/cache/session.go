package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"chat_backend/internal/database"
)

const (
	// DefaultRefreshTokenTTL Refresh Token 默认过期时间（168小时 = 7天）
	DefaultRefreshTokenTTL = 168 * time.Hour
	// MaxDevicesPerUser 每个用户最多支持的设备数量
	MaxDevicesPerUser = 5
)

// SessionInfo 会话信息
type SessionInfo struct {
	TokenID    string    `json:"token_id"`
	Token      string    `json:"token"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	DeviceID   string    `json:"device_id,omitempty"`
	DeviceName string    `json:"device_name,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
}

// SessionManager 会话管理器
type SessionManager struct {
	redis *RedisClient
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// generateTokenID 生成唯一的 Token ID
func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// StoreRefreshToken 存储 Refresh Token
func (sm *SessionManager) StoreRefreshToken(ctx context.Context, userID, username, token string, ttl time.Duration, deviceInfo map[string]string) error {
	tokenID := generateTokenID()
	now := time.Now()
	expiresAt := now.Add(ttl)

	session := SessionInfo{
		TokenID:   tokenID,
		Token:     token,
		UserID:    userID,
		Username:  username,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	// 添加设备信息
	if deviceInfo != nil {
		if deviceID, ok := deviceInfo["device_id"]; ok {
			session.DeviceID = deviceID
		}
		if deviceName, ok := deviceInfo["device_name"]; ok {
			session.DeviceName = deviceName
		}
		if userAgent, ok := deviceInfo["user_agent"]; ok {
			session.UserAgent = userAgent
		}
		if ipAddress, ok := deviceInfo["ip_address"]; ok {
			session.IPAddress = ipAddress
		}
	}

	key := KeyRefreshToken.Build(userID)

	// 使用 Hash 存储多个设备的 token
	err := sm.redis.Hash().HSetJSON(ctx, key, tokenID, session)
	if err != nil {
		return WrapCacheError("StoreRefreshToken", key, err)
	}

	// 设置整个 Hash 的过期时间
	err = sm.redis.String().Expire(ctx, key, ttl)
	if err != nil {
		return WrapCacheError("StoreRefreshToken", key, err)
	}

	// 检查并限制设备数量
	sm.cleanupOldSessions(ctx, userID)

	return nil
}

// ValidateRefreshToken 验证 Refresh Token
func (sm *SessionManager) ValidateRefreshToken(ctx context.Context, userID, token string) (*SessionInfo, error) {
	key := KeyRefreshToken.Build(userID)

	// 获取所有 token
	fields, err := sm.redis.Hash().HKeys(ctx, key)
	if err != nil {
		return nil, WrapCacheError("ValidateRefreshToken", key, err)
	}

	// 遍历查找匹配的 token
	for _, tokenID := range fields {
		var session SessionInfo
		err := sm.redis.Hash().HGetJSON(ctx, key, tokenID, &session)
		if err != nil {
			continue
		}

		if session.Token == token {
			// 检查是否过期
			if time.Now().After(session.ExpiresAt) {
				// 删除过期的 token
				_ = sm.redis.Hash().HDel(ctx, key, tokenID)
				return nil, fmt.Errorf("token expired")
			}
			return &session, nil
		}
	}

	return nil, fmt.Errorf("token not found")
}

// RevokeRefreshToken 撤销指定的 Refresh Token
func (sm *SessionManager) RevokeRefreshToken(ctx context.Context, userID, token string) error {
	key := KeyRefreshToken.Build(userID)

	// 获取所有 token
	fields, err := sm.redis.Hash().HKeys(ctx, key)
	if err != nil {
		return WrapCacheError("RevokeRefreshToken", key, err)
	}

	// 查找并删除匹配的 token
	for _, tokenID := range fields {
		var session SessionInfo
		err := sm.redis.Hash().HGetJSON(ctx, key, tokenID, &session)
		if err != nil {
			continue
		}

		if session.Token == token {
			err := sm.redis.Hash().HDel(ctx, key, tokenID)
			if err != nil {
				return WrapCacheError("RevokeRefreshToken", key, err)
			}
			return nil
		}
	}

	return fmt.Errorf("token not found")
}

// RevokeAllUserSessions 撤销用户的所有会话
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID string) error {
	key := KeyRefreshToken.Build(userID)

	err := sm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("RevokeAllUserSessions", key, err)
	}

	return nil
}

// GetUserSessions 获取用户的所有活跃会话
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]SessionInfo, error) {
	key := KeyRefreshToken.Build(userID)

	fields, err := sm.redis.Hash().HKeys(ctx, key)
	if err != nil {
		return nil, WrapCacheError("GetUserSessions", key, err)
	}

	sessions := make([]SessionInfo, 0, len(fields))
	now := time.Now()

	for _, tokenID := range fields {
		var session SessionInfo
		err := sm.redis.Hash().HGetJSON(ctx, key, tokenID, &session)
		if err != nil {
			continue
		}

		// 只返回未过期的会话
		if now.Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// cleanupOldSessions 清理旧的会话（超过设备数量限制）
func (sm *SessionManager) cleanupOldSessions(ctx context.Context, userID string) {
	key := KeyRefreshToken.Build(userID)

	fields, err := sm.redis.Hash().HKeys(ctx, key)
	if err != nil {
		return
	}

	// 如果没有超过限制，不需要清理
	if len(fields) <= MaxDevicesPerUser {
		return
	}

	// 获取所有会话并按创建时间排序
	sessions := make([]SessionInfo, 0, len(fields))
	for _, tokenID := range fields {
		var session SessionInfo
		err := sm.redis.Hash().HGetJSON(ctx, key, tokenID, &session)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	// 按创建时间升序排序（最早的在前面）
	for i := 0; i < len(sessions); i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[i].CreatedAt.After(sessions[j].CreatedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	// 删除最旧的会话
	for i := 0; i < len(sessions)-MaxDevicesPerUser; i++ {
		_ = sm.redis.Hash().HDel(ctx, key, sessions[i].TokenID)
	}
}

// RefreshSession 刷新会话（延长过期时间）
func (sm *SessionManager) RefreshSession(ctx context.Context, userID, token string, newToken string, ttl time.Duration) error {
	key := KeyRefreshToken.Build(userID)

	// 获取所有 token
	fields, err := sm.redis.Hash().HKeys(ctx, key)
	if err != nil {
		return WrapCacheError("RefreshSession", key, err)
	}

	// 查找匹配的 token
	for _, tokenID := range fields {
		var session SessionInfo
		err := sm.redis.Hash().HGetJSON(ctx, key, tokenID, &session)
		if err != nil {
			continue
		}

		if session.Token == token {
			// 更新 token 和过期时间
			session.Token = newToken
			session.ExpiresAt = time.Now().Add(ttl)

			err := sm.redis.Hash().HSetJSON(ctx, key, tokenID, session)
			if err != nil {
				return WrapCacheError("RefreshSession", key, err)
			}

			// 更新整个 Hash 的过期时间
			err = sm.redis.String().Expire(ctx, key, ttl)
			if err != nil {
				return WrapCacheError("RefreshSession", key, err)
			}

			return nil
		}
	}

	return fmt.Errorf("token not found")
}

// IsTokenValid 检查 Token 是否有效
func (sm *SessionManager) IsTokenValid(ctx context.Context, userID, token string) bool {
	session, err := sm.ValidateRefreshToken(ctx, userID, token)
	return err == nil && session != nil
}

// GetSessionCount 获取用户的活跃会话数量
func (sm *SessionManager) GetSessionCount(ctx context.Context, userID string) (int64, error) {
	key := KeyRefreshToken.Build(userID)

	count, err := sm.redis.Hash().HLen(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetSessionCount", key, err)
	}

	return count, nil
}
