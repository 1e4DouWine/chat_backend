package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat_backend/internal/database"

	"github.com/redis/go-redis/v9"
)

const (
	// UserInfoTTL 用户信息缓存过期时间（1小时）
	UserInfoTTL = 1 * time.Hour
	// UsernameToIDTTL 用户名到ID映射缓存过期时间（1小时）
	UsernameToIDTTL = 1 * time.Hour
)

// UserInfo 用户信息
type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// UserCacheManager 用户缓存管理器
type UserCacheManager struct {
	redis *RedisClient
}

// NewUserCacheManager 创建用户缓存管理器
func NewUserCacheManager() *UserCacheManager {
	return &UserCacheManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// SetUserInfo 缓存用户信息
func (ucm *UserCacheManager) SetUserInfo(ctx context.Context, userID, username, avatar string) error {
	key := KeyUser.Build(userID)

	userInfo := UserInfo{
		UserID:   userID,
		Username: username,
		Avatar:   avatar,
	}

	err := ucm.redis.String().SetJSON(ctx, key, userInfo, UserInfoTTL)
	if err != nil {
		return WrapCacheError("SetUserInfo", key, err)
	}

	return nil
}

// GetUserInfo 获取用户信息缓存
func (ucm *UserCacheManager) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	key := KeyUser.Build(userID)

	var userInfo UserInfo
	err := ucm.redis.String().GetJSON(ctx, key, &userInfo)
	if err != nil {
		if IsRedisNil(err) {
			return nil, nil
		}
		return nil, WrapCacheError("GetUserInfo", key, err)
	}

	return &userInfo, nil
}

// BatchGetUserInfo 批量获取用户信息缓存
func (ucm *UserCacheManager) BatchGetUserInfo(ctx context.Context, userIDs []string) (map[string]*UserInfo, error) {
	if len(userIDs) == 0 {
		return make(map[string]*UserInfo), nil
	}

	// 构建 keys
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = KeyUser.Build(userID)
	}

	// 使用 Pipeline 批量获取
	pipeline := ucm.redis.Pipeline().Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipeline.Get(ctx, key)
	}

	_, err := ucm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return nil, WrapCacheError("BatchGetUserInfo", "", err)
	}

	// 解析结果
	result := make(map[string]*UserInfo)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}

		var userInfo UserInfo
		if err := json.Unmarshal([]byte(val), &userInfo); err != nil {
			continue
		}

		result[userIDs[i]] = &userInfo
	}

	return result, nil
}

// DeleteUserInfo 删除用户信息缓存
func (ucm *UserCacheManager) DeleteUserInfo(ctx context.Context, userID string) error {
	key := KeyUser.Build(userID)

	err := ucm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("DeleteUserInfo", key, err)
	}

	return nil
}

// SetUsernameToID 缓存用户名到ID的映射
func (ucm *UserCacheManager) SetUsernameToID(ctx context.Context, username, userID string) error {
	key := KeyUsernameToID.Build(username)

	err := ucm.redis.String().Set(ctx, key, userID, UsernameToIDTTL)
	if err != nil {
		return WrapCacheError("SetUsernameToID", key, err)
	}

	return nil
}

// GetUserIDByUsername 通过用户名获取用户ID
func (ucm *UserCacheManager) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	key := KeyUsernameToID.Build(username)

	userID, err := ucm.redis.String().Get(ctx, key)
	if err != nil {
		if IsRedisNil(err) {
			return "", nil
		}
		return "", WrapCacheError("GetUserIDByUsername", key, err)
	}

	return userID, nil
}

// DeleteUsernameToID 删除用户名到ID的映射
func (ucm *UserCacheManager) DeleteUsernameToID(ctx context.Context, username string) error {
	key := KeyUsernameToID.Build(username)

	err := ucm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("DeleteUsernameToID", key, err)
	}

	return nil
}

// BatchSetUserInfo 批量缓存用户信息
func (ucm *UserCacheManager) BatchSetUserInfo(ctx context.Context, users []*UserInfo) error {
	if len(users) == 0 {
		return nil
	}

	// 使用 Pipeline 批量设置
	pipeline := ucm.redis.Pipeline().Pipeline()

	for _, user := range users {
		key := KeyUser.Build(user.UserID)
		data, err := json.Marshal(user)
		if err != nil {
			continue
		}
		pipeline.Set(ctx, key, data, UserInfoTTL)
	}

	_, err := ucm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return WrapCacheError("BatchSetUserInfo", "", err)
	}

	return nil
}

// UpdateUserInfo 更新用户信息缓存
func (ucm *UserCacheManager) UpdateUserInfo(ctx context.Context, userID string, updates map[string]interface{}) error {
	key := KeyUser.Build(userID)

	// 先获取现有信息
	userInfo, err := ucm.GetUserInfo(ctx, userID)
	if err != nil {
		return err
	}

	if userInfo == nil {
		return fmt.Errorf("user info not found")
	}

	// 更新字段
	if username, ok := updates["username"].(string); ok {
		userInfo.Username = username
	}
	if avatar, ok := updates["avatar"].(string); ok {
		userInfo.Avatar = avatar
	}

	// 重新设置缓存
	err = ucm.redis.String().SetJSON(ctx, key, userInfo, UserInfoTTL)
	if err != nil {
		return WrapCacheError("UpdateUserInfo", key, err)
	}

	return nil
}

// GetOrLoadUserInfo 获取或加载用户信息（Cache-Aside 模式）
func (ucm *UserCacheManager) GetOrLoadUserInfo(ctx context.Context, userID string, loader func(string) (*UserInfo, error)) (*UserInfo, error) {
	// 先尝试从缓存获取
	userInfo, err := ucm.GetUserInfo(ctx, userID)
	if err == nil && userInfo != nil {
		return userInfo, nil
	}

	// 缓存未命中，调用 loader 加载数据
	userInfo, err = loader(userID)
	if err != nil {
		return nil, err
	}

	// 将数据写入缓存
	if userInfo != nil {
		_ = ucm.SetUserInfo(ctx, userInfo.UserID, userInfo.Username, userInfo.Avatar)
	}

	return userInfo, nil
}

// GetOrLoadUserIDByUsername 获取或加载用户ID（Cache-Aside 模式）
func (ucm *UserCacheManager) GetOrLoadUserIDByUsername(ctx context.Context, username string, loader func(string) (string, error)) (string, error) {
	// 先尝试从缓存获取
	userID, err := ucm.GetUserIDByUsername(ctx, username)
	if err == nil && userID != "" {
		return userID, nil
	}

	// 缓存未命中，调用 loader 加载数据
	userID, err = loader(username)
	if err != nil {
		return "", err
	}

	// 将数据写入缓存
	if userID != "" {
		_ = ucm.SetUsernameToID(ctx, username, userID)
	}

	return userID, nil
}

// RefreshUserInfo 刷新用户信息缓存（延长过期时间）
func (ucm *UserCacheManager) RefreshUserInfo(ctx context.Context, userID string) error {
	key := KeyUser.Build(userID)

	err := ucm.redis.String().Expire(ctx, key, UserInfoTTL)
	if err != nil {
		return WrapCacheError("RefreshUserInfo", key, err)
	}

	return nil
}

// RefreshUsernameToID 刷新用户名到ID映射缓存（延长过期时间）
func (ucm *UserCacheManager) RefreshUsernameToID(ctx context.Context, username string) error {
	key := KeyUsernameToID.Build(username)

	err := ucm.redis.String().Expire(ctx, key, UsernameToIDTTL)
	if err != nil {
		return WrapCacheError("RefreshUsernameToID", key, err)
	}

	return nil
}
