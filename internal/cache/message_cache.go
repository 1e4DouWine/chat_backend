package cache

import (
	"context"
	"encoding/json"
	"time"

	"chat_backend/internal/database"

	"github.com/redis/go-redis/v9"
)

const (
	// MessageTTL 消息缓存过期时间（7天）
	MessageTTL = 7 * 24 * time.Hour
	// MessageCacheLimit 每个会话缓存的消息数量限制
	MessageCacheLimit = 100
)

// CachedMessage 缓存的消息
type CachedMessage struct {
	MessageID   string     `json:"message_id"`
	FromUserID  string     `json:"from_user_id"`
	TargetID    string     `json:"target_id"`
	Type        string     `json:"type"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"created_at"`
	FromUser    *UserInfo  `json:"from_user,omitempty"`
	TargetUser  *UserInfo  `json:"target_user,omitempty"`
	TargetGroup *GroupInfo `json:"target_group,omitempty"`
}

// GroupInfo 群组信息
type GroupInfo struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
}

// MessageCacheManager 消息缓存管理器
type MessageCacheManager struct {
	redis *RedisClient
}

// NewMessageCacheManager 创建消息缓存管理器
func NewMessageCacheManager() *MessageCacheManager {
	return &MessageCacheManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// CachePrivateMessage 缓存私聊消息
func (mcm *MessageCacheManager) CachePrivateMessage(ctx context.Context, userID, targetUserID string, message *CachedMessage) error {
	key := KeyMessagesPrivate.BuildMulti(userID, targetUserID)

	// 将消息添加到列表左侧（最新消息）
	err := mcm.redis.List().LPush(ctx, key, message.MessageID)
	if err != nil {
		return WrapCacheError("CachePrivateMessage", key, err)
	}

	// 限制列表长度
	_ = mcm.redis.List().LTrim(ctx, key, 0, int64(MessageCacheLimit)-1)

	// 缓存消息详情
	err = mcm.CacheMessageDetail(ctx, message)
	if err != nil {
		return err
	}

	// 设置过期时间
	err = mcm.redis.String().Expire(ctx, key, MessageTTL)
	if err != nil {
		return WrapCacheError("CachePrivateMessage", key, err)
	}

	return nil
}

// CacheGroupMessage 缓存群聊消息
func (mcm *MessageCacheManager) CacheGroupMessage(ctx context.Context, groupID string, message *CachedMessage) error {
	key := KeyMessagesGroup.Build(groupID)

	// 将消息添加到列表左侧（最新消息）
	err := mcm.redis.List().LPush(ctx, key, message.MessageID)
	if err != nil {
		return WrapCacheError("CacheGroupMessage", key, err)
	}

	// 限制列表长度
	_ = mcm.redis.List().LTrim(ctx, key, 0, int64(MessageCacheLimit)-1)

	// 缓存消息详情
	err = mcm.CacheMessageDetail(ctx, message)
	if err != nil {
		return err
	}

	// 设置过期时间
	err = mcm.redis.String().Expire(ctx, key, MessageTTL)
	if err != nil {
		return WrapCacheError("CacheGroupMessage", key, err)
	}

	return nil
}

// CacheMessageDetail 缓存消息详情
func (mcm *MessageCacheManager) CacheMessageDetail(ctx context.Context, message *CachedMessage) error {
	key := KeyMessage.Build(message.MessageID)

	err := mcm.redis.String().SetJSON(ctx, key, message, MessageTTL)
	if err != nil {
		return WrapCacheError("CacheMessageDetail", key, err)
	}

	return nil
}

// GetCachedPrivateMessages 获取缓存的私聊消息
func (mcm *MessageCacheManager) GetCachedPrivateMessages(ctx context.Context, userID, targetUserID string, limit int) ([]*CachedMessage, error) {
	key := KeyMessagesPrivate.BuildMulti(userID, targetUserID)

	if limit <= 0 || limit > MessageCacheLimit {
		limit = MessageCacheLimit
	}

	// 获取消息ID列表
	messageIDs, err := mcm.redis.List().LRange(ctx, key, 0, int64(limit)-1)
	if err != nil {
		if IsRedisNil(err) {
			return []*CachedMessage{}, nil
		}
		return nil, WrapCacheError("GetCachedPrivateMessages", key, err)
	}

	if len(messageIDs) == 0 {
		return []*CachedMessage{}, nil
	}

	// 批量获取消息详情
	return mcm.BatchGetMessageDetails(ctx, messageIDs)
}

// GetCachedGroupMessages 获取缓存的群聊消息
func (mcm *MessageCacheManager) GetCachedGroupMessages(ctx context.Context, groupID string, limit int) ([]*CachedMessage, error) {
	key := KeyMessagesGroup.Build(groupID)

	if limit <= 0 || limit > MessageCacheLimit {
		limit = MessageCacheLimit
	}

	// 获取消息ID列表
	messageIDs, err := mcm.redis.List().LRange(ctx, key, 0, int64(limit)-1)
	if err != nil {
		if IsRedisNil(err) {
			return []*CachedMessage{}, nil
		}
		return nil, WrapCacheError("GetCachedGroupMessages", key, err)
	}

	if len(messageIDs) == 0 {
		return []*CachedMessage{}, nil
	}

	// 批量获取消息详情
	return mcm.BatchGetMessageDetails(ctx, messageIDs)
}

// GetMessageDetail 获取单条消息详情
func (mcm *MessageCacheManager) GetMessageDetail(ctx context.Context, messageID string) (*CachedMessage, error) {
	key := KeyMessage.Build(messageID)

	var message CachedMessage
	err := mcm.redis.String().GetJSON(ctx, key, &message)
	if err != nil {
		if IsRedisNil(err) {
			return nil, nil
		}
		return nil, WrapCacheError("GetMessageDetail", key, err)
	}

	return &message, nil
}

// BatchGetMessageDetails 批量获取消息详情
func (mcm *MessageCacheManager) BatchGetMessageDetails(ctx context.Context, messageIDs []string) ([]*CachedMessage, error) {
	if len(messageIDs) == 0 {
		return []*CachedMessage{}, nil
	}

	// 构建 keys
	keys := make([]string, len(messageIDs))
	for i, messageID := range messageIDs {
		keys[i] = KeyMessage.Build(messageID)
	}

	// 使用 Pipeline 批量获取
	pipeline := mcm.redis.Pipeline().Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipeline.Get(ctx, key)
	}

	_, err := mcm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return nil, WrapCacheError("BatchGetMessageDetails", "", err)
	}

	// 解析结果
	messages := make([]*CachedMessage, 0, len(messageIDs))
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}

		var message CachedMessage
		if err := json.Unmarshal([]byte(val), &message); err != nil {
			continue
		}

		messages = append(messages, &message)
	}

	return messages, nil
}

// InvalidatePrivateMessages 使私聊消息缓存失效
func (mcm *MessageCacheManager) InvalidatePrivateMessages(ctx context.Context, userID, targetUserID string) error {
	key := KeyMessagesPrivate.BuildMulti(userID, targetUserID)

	err := mcm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidatePrivateMessages", key, err)
	}

	return nil
}

// InvalidateGroupMessages 使群聊消息缓存失效
func (mcm *MessageCacheManager) InvalidateGroupMessages(ctx context.Context, groupID string) error {
	key := KeyMessagesGroup.Build(groupID)

	err := mcm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidateGroupMessages", key, err)
	}

	return nil
}

// InvalidateMessage 使单条消息缓存失效
func (mcm *MessageCacheManager) InvalidateMessage(ctx context.Context, messageID string) error {
	key := KeyMessage.Build(messageID)

	err := mcm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidateMessage", key, err)
	}

	return nil
}

// GetPrivateMessageCount 获取私聊消息缓存数量
func (mcm *MessageCacheManager) GetPrivateMessageCount(ctx context.Context, userID, targetUserID string) (int64, error) {
	key := KeyMessagesPrivate.BuildMulti(userID, targetUserID)

	count, err := mcm.redis.List().LLen(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetPrivateMessageCount", key, err)
	}

	return count, nil
}

// GetGroupMessageCount 获取群聊消息缓存数量
func (mcm *MessageCacheManager) GetGroupMessageCount(ctx context.Context, groupID string) (int64, error) {
	key := KeyMessagesGroup.Build(groupID)

	count, err := mcm.redis.List().LLen(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetGroupMessageCount", key, err)
	}

	return count, nil
}

// RefreshPrivateMessages 刷新私聊消息缓存（延长过期时间）
func (mcm *MessageCacheManager) RefreshPrivateMessages(ctx context.Context, userID, targetUserID string) error {
	key := KeyMessagesPrivate.BuildMulti(userID, targetUserID)

	err := mcm.redis.String().Expire(ctx, key, MessageTTL)
	if err != nil {
		return WrapCacheError("RefreshPrivateMessages", key, err)
	}

	return nil
}

// RefreshGroupMessages 刷新群聊消息缓存（延长过期时间）
func (mcm *MessageCacheManager) RefreshGroupMessages(ctx context.Context, groupID string) error {
	key := KeyMessagesGroup.Build(groupID)

	err := mcm.redis.String().Expire(ctx, key, MessageTTL)
	if err != nil {
		return WrapCacheError("RefreshGroupMessages", key, err)
	}

	return nil
}
