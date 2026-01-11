package cache

import (
	"context"
	"encoding/json"
	"time"

	"chat_backend/internal/database"
)

const (
	// ConversationTTL 会话列表缓存过期时间（5分钟）
	ConversationTTL = 5 * time.Minute
)

// PrivateConversation 私聊会话
type PrivateConversation struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Avatar      string    `json:"avatar"`
	LastContent string    `json:"last_content"`
	LastTime    time.Time `json:"last_time"`
}

// GroupConversation 群聊会话
type GroupConversation struct {
	GroupID        string    `json:"group_id"`
	GroupName      string    `json:"group_name"`
	LastContent    string    `json:"last_content"`
	LastTime       time.Time `json:"last_time"`
	LastSenderID   string    `json:"last_sender_id"`
	LastSenderName string    `json:"last_sender_name"`
}

// ConversationCacheManager 会话缓存管理器
type ConversationCacheManager struct {
	redis *RedisClient
}

// NewConversationCacheManager 创建会话缓存管理器
func NewConversationCacheManager() *ConversationCacheManager {
	return &ConversationCacheManager{
		redis: NewRedisClient(database.GetRedis()),
	}
}

// SetPrivateConversation 缓存私聊会话
func (ccm *ConversationCacheManager) SetPrivateConversation(ctx context.Context, userID string, conv *PrivateConversation) error {
	key := KeyConversations.Build(userID)
	field := "private:" + conv.UserID

	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}

	err = ccm.redis.Hash().HSet(ctx, key, field, data)
	if err != nil {
		return WrapCacheError("SetPrivateConversation", key, err)
	}

	// 设置过期时间
	err = ccm.redis.String().Expire(ctx, key, ConversationTTL)
	if err != nil {
		return WrapCacheError("SetPrivateConversation", key, err)
	}

	return nil
}

// SetGroupConversation 缓存群聊会话
func (ccm *ConversationCacheManager) SetGroupConversation(ctx context.Context, userID string, conv *GroupConversation) error {
	key := KeyConversations.Build(userID)
	field := "group:" + conv.GroupID

	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}

	err = ccm.redis.Hash().HSet(ctx, key, field, data)
	if err != nil {
		return WrapCacheError("SetGroupConversation", key, err)
	}

	// 设置过期时间
	err = ccm.redis.String().Expire(ctx, key, ConversationTTL)
	if err != nil {
		return WrapCacheError("SetGroupConversation", key, err)
	}

	return nil
}

// BatchSetPrivateConversations 批量缓存私聊会话
func (ccm *ConversationCacheManager) BatchSetPrivateConversations(ctx context.Context, userID string, convs []*PrivateConversation) error {
	if len(convs) == 0 {
		return nil
	}

	key := KeyConversations.Build(userID)

	// 使用 Pipeline 批量设置
	pipeline := ccm.redis.Pipeline().Pipeline()

	for _, conv := range convs {
		field := "private:" + conv.UserID
		data, err := json.Marshal(conv)
		if err != nil {
			continue
		}
		pipeline.HSet(ctx, key, field, data)
	}

	_, err := ccm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return WrapCacheError("BatchSetPrivateConversations", key, err)
	}

	// 设置过期时间
	err = ccm.redis.String().Expire(ctx, key, ConversationTTL)
	if err != nil {
		return WrapCacheError("BatchSetPrivateConversations", key, err)
	}

	return nil
}

// BatchSetGroupConversations 批量缓存群聊会话
func (ccm *ConversationCacheManager) BatchSetGroupConversations(ctx context.Context, userID string, convs []*GroupConversation) error {
	if len(convs) == 0 {
		return nil
	}

	key := KeyConversations.Build(userID)

	// 使用 Pipeline 批量设置
	pipeline := ccm.redis.Pipeline().Pipeline()

	for _, conv := range convs {
		field := "group:" + conv.GroupID
		data, err := json.Marshal(conv)
		if err != nil {
			continue
		}
		pipeline.HSet(ctx, key, field, data)
	}

	_, err := ccm.redis.Pipeline().Exec(ctx, pipeline)
	if err != nil {
		return WrapCacheError("BatchSetGroupConversations", key, err)
	}

	// 设置过期时间
	err = ccm.redis.String().Expire(ctx, key, ConversationTTL)
	if err != nil {
		return WrapCacheError("BatchSetGroupConversations", key, err)
	}

	return nil
}

// GetPrivateConversation 获取私聊会话缓存
func (ccm *ConversationCacheManager) GetPrivateConversation(ctx context.Context, userID, partnerID string) (*PrivateConversation, error) {
	key := KeyConversations.Build(userID)
	field := "private:" + partnerID

	var conv PrivateConversation
	err := ccm.redis.Hash().HGetJSON(ctx, key, field, &conv)
	if err != nil {
		if IsRedisNil(err) {
			return nil, nil
		}
		return nil, WrapCacheError("GetPrivateConversation", key, err)
	}

	return &conv, nil
}

// GetGroupConversation 获取群聊会话缓存
func (ccm *ConversationCacheManager) GetGroupConversation(ctx context.Context, userID, groupID string) (*GroupConversation, error) {
	key := KeyConversations.Build(userID)
	field := "group:" + groupID

	var conv GroupConversation
	err := ccm.redis.Hash().HGetJSON(ctx, key, field, &conv)
	if err != nil {
		if IsRedisNil(err) {
			return nil, nil
		}
		return nil, WrapCacheError("GetGroupConversation", key, err)
	}

	return &conv, nil
}

// GetAllConversations 获取所有会话缓存
func (ccm *ConversationCacheManager) GetAllConversations(ctx context.Context, userID string) (map[string]interface{}, error) {
	key := KeyConversations.Build(userID)

	fields, err := ccm.redis.Hash().HGetAll(ctx, key)
	if err != nil {
		if IsRedisNil(err) {
			return make(map[string]interface{}), nil
		}
		return nil, WrapCacheError("GetAllConversations", key, err)
	}

	if len(fields) == 0 {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})
	for field, data := range fields {
		if len(field) > 8 && field[:8] == "private:" {
			var conv PrivateConversation
			if err := json.Unmarshal([]byte(data), &conv); err == nil {
				result[field] = conv
			}
		} else if len(field) > 6 && field[:6] == "group:" {
			var conv GroupConversation
			if err := json.Unmarshal([]byte(data), &conv); err == nil {
				result[field] = conv
			}
		}
	}

	return result, nil
}

// InvalidateConversation 使会话缓存失效
func (ccm *ConversationCacheManager) InvalidateConversation(ctx context.Context, userID string) error {
	key := KeyConversations.Build(userID)

	err := ccm.redis.String().Del(ctx, key)
	if err != nil {
		return WrapCacheError("InvalidateConversation", key, err)
	}

	return nil
}

// InvalidatePrivateConversation 使私聊会话缓存失效
func (ccm *ConversationCacheManager) InvalidatePrivateConversation(ctx context.Context, userID, partnerID string) error {
	key := KeyConversations.Build(userID)
	field := "private:" + partnerID

	err := ccm.redis.Hash().HDel(ctx, key, field)
	if err != nil {
		return WrapCacheError("InvalidatePrivateConversation", key, err)
	}

	return nil
}

// InvalidateGroupConversation 使群聊会话缓存失效
func (ccm *ConversationCacheManager) InvalidateGroupConversation(ctx context.Context, userID, groupID string) error {
	key := KeyConversations.Build(userID)
	field := "group:" + groupID

	err := ccm.redis.Hash().HDel(ctx, key, field)
	if err != nil {
		return WrapCacheError("InvalidateGroupConversation", key, err)
	}

	return nil
}

// RefreshConversation 刷新会话缓存（延长过期时间）
func (ccm *ConversationCacheManager) RefreshConversation(ctx context.Context, userID string) error {
	key := KeyConversations.Build(userID)

	err := ccm.redis.String().Expire(ctx, key, ConversationTTL)
	if err != nil {
		return WrapCacheError("RefreshConversation", key, err)
	}

	return nil
}

// UpdatePrivateConversation 更新私聊会话缓存
func (ccm *ConversationCacheManager) UpdatePrivateConversation(ctx context.Context, userID string, conv *PrivateConversation) error {
	return ccm.SetPrivateConversation(ctx, userID, conv)
}

// UpdateGroupConversation 更新群聊会话缓存
func (ccm *ConversationCacheManager) UpdateGroupConversation(ctx context.Context, userID string, conv *GroupConversation) error {
	return ccm.SetGroupConversation(ctx, userID, conv)
}

// GetConversationCount 获取会话数量
func (ccm *ConversationCacheManager) GetConversationCount(ctx context.Context, userID string) (int64, error) {
	key := KeyConversations.Build(userID)

	count, err := ccm.redis.Hash().HLen(ctx, key)
	if err != nil {
		return 0, WrapCacheError("GetConversationCount", key, err)
	}

	return count, nil
}
