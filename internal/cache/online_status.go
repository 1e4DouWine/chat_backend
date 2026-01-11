package cache

import (
	"context"
	"fmt"
	"time"

	"chat_backend/internal/database"

	"github.com/redis/go-redis/v9"
)

const (
	// OnlineStatusTTL 在线状态过期时间（60秒）
	OnlineStatusTTL = 60 * time.Second
	// OnlineHeartbeatInterval 心跳间隔（30秒）
	OnlineHeartbeatInterval = 30 * time.Second
)

// OnlineStatusManager 在线状态管理器
type OnlineStatusManager struct {
	redis *RedisClient
}

// NewOnlineStatusManager 创建在线状态管理器
func NewOnlineStatusManager() *OnlineStatusManager {
	return &OnlineStatusManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// SetOnline 设置用户在线状态
func (osm *OnlineStatusManager) SetOnline(ctx context.Context, userID string) error {
	key := KeyOnlineUsers.Build()
	now := float64(time.Now().Unix())

	// 使用 Sorted Set，member 为 userID，score 为最后心跳时间戳
	err := osm.redis.SortedSet().ZAdd(ctx, key, redis.Z{
		Score:  now,
		Member: userID,
	})
	if err != nil {
		return WrapCacheError("SetOnline", key, err)
	}

	return nil
}

// SetOffline 设置用户离线状态
func (osm *OnlineStatusManager) SetOffline(ctx context.Context, userID string) error {
	key := KeyOnlineUsers.Build()

	err := osm.redis.SortedSet().ZRem(ctx, key, userID)
	if err != nil {
		return WrapCacheError("SetOffline", key, err)
	}

	return nil
}

// IsOnline 检查用户是否在线
func (osm *OnlineStatusManager) IsOnline(ctx context.Context, userID string) (bool, error) {
	key := KeyOnlineUsers.Build()

	// 检查用户是否在在线用户集合中
	count, err := osm.redis.SortedSet().ZCard(ctx, key)
	if err != nil {
		return false, WrapCacheError("IsOnline", key, err)
	}

	if count == 0 {
		return false, nil
	}

	// 获取用户的分数（心跳时间）
	score, err := osm.redis.SortedSet().ZScore(ctx, key, userID)
	if err != nil {
		// 如果用户不存在，返回 false
		if IsRedisNil(err) {
			return false, nil
		}
		return false, WrapCacheError("IsOnline", key, err)
	}

	// 检查心跳是否过期
	now := float64(time.Now().Unix())
	if now-score > float64(OnlineStatusTTL.Seconds()) {
		// 心跳过期，清理该用户
		_ = osm.SetOffline(ctx, userID)
		return false, nil
	}

	return true, nil
}

// Heartbeat 更新用户心跳
func (osm *OnlineStatusManager) Heartbeat(ctx context.Context, userID string) error {
	key := KeyOnlineUsers.Build()
	now := float64(time.Now().Unix())

	err := osm.redis.SortedSet().ZAdd(ctx, key, redis.Z{
		Score:  now,
		Member: userID,
	})
	if err != nil {
		return WrapCacheError("Heartbeat", key, err)
	}

	return nil
}

// GetOnlineUserCount 获取在线用户数量
func (osm *OnlineStatusManager) GetOnlineUserCount(ctx context.Context) (int64, error) {
	key := KeyOnlineUsers.Build()

	// 先清理过期的用户
	_ = osm.CleanupExpiredUsers(ctx)

	count, err := osm.redis.SortedSet().ZCard(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetOnlineUserCount", key, err)
	}

	return count, nil
}

// GetOnlineUserIDs 获取所有在线用户 ID 列表
func (osm *OnlineStatusManager) GetOnlineUserIDs(ctx context.Context) ([]string, error) {
	key := KeyOnlineUsers.Build()

	// 先清理过期的用户
	_ = osm.CleanupExpiredUsers(ctx)

	// 获取所有成员
	userIDs, err := osm.redis.SortedSet().ZRange(ctx, key, 0, -1)
	if err != nil {
		return nil, WrapCacheError("GetOnlineUserIDs", key, err)
	}

	return userIDs, nil
}

// GetOnlineUsersWithScores 获取在线用户及其心跳时间
func (osm *OnlineStatusManager) GetOnlineUsersWithScores(ctx context.Context) ([]redis.Z, error) {
	key := KeyOnlineUsers.Build()

	// 先清理过期的用户
	_ = osm.CleanupExpiredUsers(ctx)

	// 获取所有成员及其分数
	members, err := osm.redis.SortedSet().ZRangeWithScores(ctx, key, 0, -1)
	if err != nil {
		return nil, WrapCacheError("GetOnlineUsersWithScores", key, err)
	}

	return members, nil
}

// CleanupExpiredUsers 清理过期的在线用户
func (osm *OnlineStatusManager) CleanupExpiredUsers(ctx context.Context) error {
	key := KeyOnlineUsers.Build()
	now := float64(time.Now().Unix())
	expireThreshold := now - float64(OnlineStatusTTL.Seconds())

	// 删除分数小于阈值的用户（即心跳过期的用户）
	count, err := osm.redis.SortedSet().ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", expireThreshold))
	if err != nil {
		return WrapCacheError("CleanupExpiredUsers", key, err)
	}

	if count > 0 {
		// 记录清理的用户数量
		_ = fmt.Sprintf("Cleaned up %d expired online users", count)
	}

	return nil
}

// BatchSetOnline 批量设置用户在线状态
func (osm *OnlineStatusManager) BatchSetOnline(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	key := KeyOnlineUsers.Build()
	now := float64(time.Now().Unix())

	// 构建成员列表
	members := make([]redis.Z, len(userIDs))
	for i, userID := range userIDs {
		members[i] = redis.Z{
			Score:  now,
			Member: userID,
		}
	}

	err := osm.redis.SortedSet().ZAdd(ctx, key, members...)
	if err != nil {
		return WrapCacheError("BatchSetOnline", key, err)
	}

	return nil
}

// BatchSetOffline 批量设置用户离线状态
func (osm *OnlineStatusManager) BatchSetOffline(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	key := KeyOnlineUsers.Build()

	// 转换为 interface{} 类型
	members := make([]interface{}, len(userIDs))
	for i, userID := range userIDs {
		members[i] = userID
	}

	err := osm.redis.SortedSet().ZRem(ctx, key, members...)
	if err != nil {
		return WrapCacheError("BatchSetOffline", key, err)
	}

	return nil
}

// GetLastHeartbeat 获取用户最后心跳时间
func (osm *OnlineStatusManager) GetLastHeartbeat(ctx context.Context, userID string) (time.Time, error) {
	key := KeyOnlineUsers.Build()

	score, err := osm.redis.SortedSet().ZScore(ctx, key, userID)
	if err != nil {
		if IsRedisNil(err) {
			return time.Time{}, nil
		}
		return time.Time{}, WrapCacheError("GetLastHeartbeat", key, err)
	}

	return time.Unix(int64(score), 0), nil
}

// IsUserOnlineWithinDuration 检查用户在指定时间内是否在线
func (osm *OnlineStatusManager) IsUserOnlineWithinDuration(ctx context.Context, userID string, duration time.Duration) (bool, error) {
	key := KeyOnlineUsers.Build()

	score, err := osm.redis.SortedSet().ZScore(ctx, key, userID)
	if err != nil {
		if IsRedisNil(err) {
			return false, nil
		}
		return false, WrapCacheError("IsUserOnlineWithinDuration", key, err)
	}

	now := float64(time.Now().Unix())
	if now-score > float64(duration.Seconds()) {
		return false, nil
	}

	return true, nil
}

// GetOnlineUserIDsByRange 获取指定时间范围内的在线用户
func (osm *OnlineStatusManager) GetOnlineUserIDsByRange(ctx context.Context, start, end time.Time) ([]string, error) {
	key := KeyOnlineUsers.Build()

	// 先清理过期的用户
	_ = osm.CleanupExpiredUsers(ctx)

	// 获取指定范围内的用户
	userIDs, err := osm.redis.SortedSet().ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", float64(start.Unix())),
		Max: fmt.Sprintf("%f", float64(end.Unix())),
	})
	if err != nil {
		return nil, WrapCacheError("GetOnlineUserIDsByRange", key, err)
	}

	return userIDs, nil
}

// ClearAllOnlineUsers 清空所有在线用户（慎用）
func (osm *OnlineStatusManager) ClearAllOnlineUsers(ctx context.Context) error {
	key := KeyOnlineUsers.Build()

	err := osm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("ClearAllOnlineUsers", key, err)
	}

	return nil
}
