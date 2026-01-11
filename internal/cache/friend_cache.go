package cache

import (
	"context"
	"encoding/json"
	"time"

	"chat_backend/internal/database"
)

const (
	// FriendListTTL 好友列表缓存过期时间（30分钟）
	FriendListTTL = 30 * time.Minute
	// FriendRequestTTL 好友申请缓存过期时间（5分钟）
	FriendRequestTTL = 5 * time.Minute
)

// FriendInfo 好友信息
type FriendInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
	CreateAt int64  `json:"create_at"`
}

// FriendCacheManager 好友缓存管理器
type FriendCacheManager struct {
	redis *RedisClient
}

// NewFriendCacheManager 创建好友缓存管理器
func NewFriendCacheManager() *FriendCacheManager {
	return &FriendCacheManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// AddFriend 添加好友到缓存
func (fcm *FriendCacheManager) AddFriend(ctx context.Context, userID, friendID string) error {
	key := KeyFriends.Build(userID)

	err := fcm.redis.Set().SAdd(ctx, key, friendID)
	if err != nil {
		return WrapCacheError("AddFriend", key, err)
	}

	// 设置过期时间
	err = fcm.redis.String().Expire(ctx, key, FriendListTTL)
	if err != nil {
		return WrapCacheError("AddFriend", key, err)
	}

	return nil
}

// RemoveFriend 从缓存中移除好友
func (fcm *FriendCacheManager) RemoveFriend(ctx context.Context, userID, friendID string) error {
	key := KeyFriends.Build(userID)

	err := fcm.redis.Set().SRem(ctx, key, friendID)
	if err != nil {
		return WrapCacheError("RemoveFriend", key, err)
	}

	return nil
}

// IsFriend 检查是否是好友
func (fcm *FriendCacheManager) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	key := KeyFriends.Build(userID)

	isMember, err := fcm.redis.Set().SIsMember(ctx, key, friendID)
	if err != nil {
		return false, WrapCacheError("IsFriend", key, err)
	}

	return isMember, nil
}

// GetFriendList 获取好友列表
func (fcm *FriendCacheManager) GetFriendList(ctx context.Context, userID string) ([]string, error) {
	key := KeyFriends.Build(userID)

	friendIDs, err := fcm.redis.Set().SMembers(ctx, key)
	if err != nil {
		if IsRedisNil(err) {
			return []string{}, nil
		}
		return nil, WrapCacheError("GetFriendList", key, err)
	}

	return friendIDs, nil
}

// GetFriendCount 获取好友数量
func (fcm *FriendCacheManager) GetFriendCount(ctx context.Context, userID string) (int64, error) {
	key := KeyFriends.Build(userID)

	count, err := fcm.redis.Set().SCard(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetFriendCount", key, err)
	}

	return count, nil
}

// BatchAddFriends 批量添加好友
func (fcm *FriendCacheManager) BatchAddFriends(ctx context.Context, userID string, friendIDs []string) error {
	if len(friendIDs) == 0 {
		return nil
	}

	key := KeyFriends.Build(userID)

	// 转换为 interface{} 类型
	members := make([]interface{}, len(friendIDs))
	for i, friendID := range friendIDs {
		members[i] = friendID
	}

	err := fcm.redis.Set().SAdd(ctx, key, members...)
	if err != nil {
		return WrapCacheError("BatchAddFriends", key, err)
	}

	// 设置过期时间
	err = fcm.redis.String().Expire(ctx, key, FriendListTTL)
	if err != nil {
		return WrapCacheError("BatchAddFriends", key, err)
	}

	return nil
}

// BatchRemoveFriends 批量移除好友
func (fcm *FriendCacheManager) BatchRemoveFriends(ctx context.Context, userID string, friendIDs []string) error {
	if len(friendIDs) == 0 {
		return nil
	}

	key := KeyFriends.Build(userID)

	// 转换为 interface{} 类型
	members := make([]interface{}, len(friendIDs))
	for i, friendID := range friendIDs {
		members[i] = friendID
	}

	err := fcm.redis.Set().SRem(ctx, key, members...)
	if err != nil {
		return WrapCacheError("BatchRemoveFriends", key, err)
	}

	return nil
}

// InvalidateFriendList 使好友列表缓存失效
func (fcm *FriendCacheManager) InvalidateFriendList(ctx context.Context, userID string) error {
	key := KeyFriends.Build(userID)

	err := fcm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidateFriendList", key, err)
	}

	return nil
}

// RefreshFriendList 刷新好友列表缓存（延长过期时间）
func (fcm *FriendCacheManager) RefreshFriendList(ctx context.Context, userID string) error {
	key := KeyFriends.Build(userID)

	err := fcm.redis.String().Expire(ctx, key, FriendListTTL)
	if err != nil {
		return WrapCacheError("RefreshFriendList", key, err)
	}

	return nil
}

// AddFriendRequest 添加好友申请到缓存
func (fcm *FriendCacheManager) AddFriendRequest(ctx context.Context, userID string, request *FriendInfo) error {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = fcm.redis.List().LPush(ctx, key, data)
	if err != nil {
		return WrapCacheError("AddFriendRequest", key, err)
	}

	// 设置过期时间
	err = fcm.redis.String().Expire(ctx, key, FriendRequestTTL)
	if err != nil {
		return WrapCacheError("AddFriendRequest", key, err)
	}

	return nil
}

// GetFriendRequests 获取好友申请列表
func (fcm *FriendCacheManager) GetFriendRequests(ctx context.Context, userID string) ([]*FriendInfo, error) {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	requests, err := fcm.redis.List().LRange(ctx, key, 0, -1)
	if err != nil {
		if IsRedisNil(err) {
			return []*FriendInfo{}, nil
		}
		return nil, WrapCacheError("GetFriendRequests", key, err)
	}

	if len(requests) == 0 {
		return []*FriendInfo{}, nil
	}

	result := make([]*FriendInfo, 0, len(requests))
	for _, req := range requests {
		var friendInfo FriendInfo
		if err := json.Unmarshal([]byte(req), &friendInfo); err == nil {
			result = append(result, &friendInfo)
		}
	}

	return result, nil
}

// RemoveFriendRequest 从缓存中移除好友申请
func (fcm *FriendCacheManager) RemoveFriendRequest(ctx context.Context, userID, senderID string) error {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	// 获取所有申请
	requests, err := fcm.redis.List().LRange(ctx, key, 0, -1)
	if err != nil {
		if IsRedisNil(err) {
			return nil
		}
		return WrapCacheError("RemoveFriendRequest", key, err)
	}

	// 查找并删除匹配的申请
	for _, req := range requests {
		var friendInfo FriendInfo
		if err := json.Unmarshal([]byte(req), &friendInfo); err == nil {
			if friendInfo.UserID == senderID {
				// 使用 LREM 删除
				_, _ = fcm.redis.List().LRem(ctx, key, 0, req)
				break
			}
		}
	}

	return nil
}

// InvalidateFriendRequests 使好友申请缓存失效
func (fcm *FriendCacheManager) InvalidateFriendRequests(ctx context.Context, userID string) error {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	err := fcm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidateFriendRequests", key, err)
	}

	return nil
}

// RefreshFriendRequests 刷新好友申请缓存（延长过期时间）
func (fcm *FriendCacheManager) RefreshFriendRequests(ctx context.Context, userID string) error {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	err := fcm.redis.String().Expire(ctx, key, FriendRequestTTL)
	if err != nil {
		return WrapCacheError("RefreshFriendRequests", key, err)
	}

	return nil
}

// BatchAddFriendRequests 批量添加好友申请
func (fcm *FriendCacheManager) BatchAddFriendRequests(ctx context.Context, userID string, requests []*FriendInfo) error {
	if len(requests) == 0 {
		return nil
	}

	key := KeyFriendRequests.BuildMulti("pending", userID)

	// 使用 Pipeline 批量添加
	pipeline := fcm.redis.Pipeline().Pipeline()

	for _, req := range requests {
		data, err := json.Marshal(req)
		if err != nil {
			continue
		}
		pipeline.LPush(ctx, key, data)
	}

	_, err := fcm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return WrapCacheError("BatchAddFriendRequests", key, err)
	}

	// 设置过期时间
	err = fcm.redis.String().Expire(ctx, key, FriendRequestTTL)
	if err != nil {
		return WrapCacheError("BatchAddFriendRequests", key, err)
	}

	return nil
}

// GetFriendRequestCount 获取好友申请数量
func (fcm *FriendCacheManager) GetFriendRequestCount(ctx context.Context, userID string) (int64, error) {
	key := KeyFriendRequests.BuildMulti("pending", userID)

	count, err := fcm.redis.List().LLen(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetFriendRequestCount", key, err)
	}

	return count, nil
}
