package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客户端包装器
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

// GetClient 获取原始 Redis 客户端
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// StringOperations 字符串操作
type StringOperations struct {
	client *redis.Client
}

// String 返回字符串操作
func (r *RedisClient) String() *StringOperations {
	return &StringOperations{client: r.client}
}

// Set 设置字符串值
func (s *StringOperations) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取字符串值
func (s *StringOperations) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// GetJSON 获取并解析 JSON
func (s *StringOperations) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// SetJSON 设置 JSON 值
func (s *StringOperations) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, expiration).Err()
}

// Del 删除键
func (s *StringOperations) Del(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (s *StringOperations) Exists(ctx context.Context, keys ...string) (int64, error) {
	return s.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (s *StringOperations) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return s.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (s *StringOperations) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.client.TTL(ctx, key).Result()
}

// Incr 递增
func (s *StringOperations) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func (s *StringOperations) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.IncrBy(ctx, key, value).Result()
}

// HashOperations Hash 操作
type HashOperations struct {
	client *redis.Client
}

// Hash 返回 Hash 操作
func (r *RedisClient) Hash() *HashOperations {
	return &HashOperations{client: r.client}
}

// HSet 设置 Hash 字段
func (h *HashOperations) HSet(ctx context.Context, key string, values ...interface{}) error {
	return h.client.HSet(ctx, key, values...).Err()
}

// HGet 获取 Hash 字段值
func (h *HashOperations) HGet(ctx context.Context, key, field string) (string, error) {
	return h.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有 Hash 字段
func (h *HashOperations) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return h.client.HGetAll(ctx, key).Result()
}

// HGetJSON 获取并解析 JSON 字段
func (h *HashOperations) HGetJSON(ctx context.Context, key, field string, dest interface{}) error {
	val, err := h.client.HGet(ctx, key, field).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// HSetJSON 设置 JSON 字段
func (h *HashOperations) HSetJSON(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return h.client.HSet(ctx, key, field, data).Err()
}

// HDel 删除 Hash 字段
func (h *HashOperations) HDel(ctx context.Context, key string, fields ...string) error {
	return h.client.HDel(ctx, key, fields...).Err()
}

// HExists 检查 Hash 字段是否存在
func (h *HashOperations) HExists(ctx context.Context, key, field string) (bool, error) {
	return h.client.HExists(ctx, key, field).Result()
}

// HKeys 获取所有 Hash 字段名
func (h *HashOperations) HKeys(ctx context.Context, key string) ([]string, error) {
	return h.client.HKeys(ctx, key).Result()
}

// HVals 获取所有 Hash 字段值
func (h *HashOperations) HVals(ctx context.Context, key string) ([]string, error) {
	return h.client.HVals(ctx, key).Result()
}

// HLen 获取 Hash 字段数量
func (h *HashOperations) HLen(ctx context.Context, key string) (int64, error) {
	return h.client.HLen(ctx, key).Result()
}

// ListOperations List 操作
type ListOperations struct {
	client *redis.Client
}

// List 返回 List 操作
func (r *RedisClient) List() *ListOperations {
	return &ListOperations{client: r.client}
}

// LPush 从左侧推入
func (l *ListOperations) LPush(ctx context.Context, key string, values ...interface{}) error {
	return l.client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入
func (l *ListOperations) RPush(ctx context.Context, key string, values ...interface{}) error {
	return l.client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出
func (l *ListOperations) LPop(ctx context.Context, key string) (string, error) {
	return l.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出
func (l *ListOperations) RPop(ctx context.Context, key string) (string, error) {
	return l.client.RPop(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (l *ListOperations) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return l.client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (l *ListOperations) LLen(ctx context.Context, key string) (int64, error) {
	return l.client.LLen(ctx, key).Result()
}

// LTrim 修剪列表
func (l *ListOperations) LTrim(ctx context.Context, key string, start, stop int64) error {
	return l.client.LTrim(ctx, key, start, stop).Err()
}

// LRem 从列表中移除指定值的元素
func (l *ListOperations) LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error) {
	return l.client.LRem(ctx, key, count, value).Result()
}

// SetOperations Set 操作
type SetOperations struct {
	client *redis.Client
}

// Set 返回 Set 操作
func (r *RedisClient) Set() *SetOperations {
	return &SetOperations{client: r.client}
}

// SAdd 添加成员到集合
func (s *SetOperations) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return s.client.SAdd(ctx, key, members...).Err()
}

// SRem 从集合中移除成员
func (s *SetOperations) SRem(ctx context.Context, key string, members ...interface{}) error {
	return s.client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (s *SetOperations) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.client.SMembers(ctx, key).Result()
}

// SIsMember 检查成员是否在集合中
func (s *SetOperations) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return s.client.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合成员数量
func (s *SetOperations) SCard(ctx context.Context, key string) (int64, error) {
	return s.client.SCard(ctx, key).Result()
}

// SortedSetOperations Sorted Set 操作
type SortedSetOperations struct {
	client *redis.Client
}

// SortedSet 返回 Sorted Set 操作
func (r *RedisClient) SortedSet() *SortedSetOperations {
	return &SortedSetOperations{client: r.client}
}

// ZAdd 添加成员到有序集合
func (z *SortedSetOperations) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return z.client.ZAdd(ctx, key, members...).Err()
}

// ZRem 从有序集合中移除成员
func (z *SortedSetOperations) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return z.client.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (z *SortedSetOperations) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return z.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员（带分数）
func (z *SortedSetOperations) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return z.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZScore 获取成员的分数
func (z *SortedSetOperations) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return z.client.ZScore(ctx, key, member).Result()
}

// ZCard 获取有序集合成员数量
func (z *SortedSetOperations) ZCard(ctx context.Context, key string) (int64, error) {
	return z.client.ZCard(ctx, key).Result()
}

// ZCount 获取分数范围内的成员数量
func (z *SortedSetOperations) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZCount(ctx, key, min, max).Result()
}

// ZRemRangeByScore 按分数范围移除成员
func (z *SortedSetOperations) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZRemRangeByScore(ctx, key, min, max).Result()
}

// ZRangeByScore 按分数范围获取成员
func (z *SortedSetOperations) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRangeByScoreWithScores 按分数范围获取成员（带分数）
func (z *SortedSetOperations) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return z.client.ZRangeByScoreWithScores(ctx, key, opt).Result()
}

// PipelineOperations Pipeline 操作
type PipelineOperations struct {
	client *redis.Client
}

// Pipeline 返回 Pipeline 操作
func (r *RedisClient) Pipeline() *PipelineOperations {
	return &PipelineOperations{client: r.client}
}

// Pipeline 创建 Pipeline
func (p *PipelineOperations) Pipeline() redis.Pipeliner {
	return p.client.Pipeline()
}

// Exec 执行 Pipeline
func (p *PipelineOperations) Exec(ctx context.Context, pipeliner redis.Pipeliner) ([]redis.Cmder, error) {
	return pipeliner.Exec(ctx)
}

// TransactionOperations 事务操作
type TransactionOperations struct {
	client *redis.Client
}

// Transaction 返回事务操作
func (r *RedisClient) Transaction() *TransactionOperations {
	return &TransactionOperations{client: r.client}
}

// Watch 监视键
func (t *TransactionOperations) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return t.client.Watch(ctx, fn, keys...)
}

// TxPipeline 创建事务 Pipeline
func (t *TransactionOperations) TxPipeline() redis.Pipeliner {
	return t.client.TxPipeline()
}

// CacheKey 缓存键生成器
type CacheKey struct {
	prefix string
}

// NewCacheKey 创建缓存键生成器
func NewCacheKey(prefix string) *CacheKey {
	return &CacheKey{prefix: prefix}
}

// Build 构建缓存键
func (ck *CacheKey) Build(parts ...string) string {
	if len(parts) == 0 {
		return ck.prefix
	}
	return fmt.Sprintf("%s:%s", ck.prefix, parts[0])
}

// BuildMulti 构建多部分缓存键
func (ck *CacheKey) BuildMulti(parts ...string) string {
	if len(parts) == 0 {
		return ck.prefix
	}
	return fmt.Sprintf("%s:%s", ck.prefix, joinParts(parts...))
}

// joinParts 连接多个部分
func joinParts(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}

// 预定义的缓存键前缀
const (
	KeyPrefixRefreshToken    = "refresh_token"
	KeyPrefixOnlineUsers     = "online_users"
	KeyPrefixUser            = "user"
	KeyPrefixUsernameToID    = "username_to_id"
	KeyPrefixMessage         = "message"
	KeyPrefixMessagesPrivate = "messages:private"
	KeyPrefixMessagesGroup   = "messages:group"
	KeyPrefixConversations   = "conversations"
	KeyPrefixFriends         = "friends"
	KeyPrefixFriendRequests  = "friend_requests"
	KeyPrefixRateLimit       = "rate_limit"
)

// 预定义的缓存键生成器
var (
	KeyRefreshToken    = NewCacheKey(KeyPrefixRefreshToken)
	KeyOnlineUsers     = NewCacheKey(KeyPrefixOnlineUsers)
	KeyUser            = NewCacheKey(KeyPrefixUser)
	KeyUsernameToID    = NewCacheKey(KeyPrefixUsernameToID)
	KeyMessage         = NewCacheKey(KeyPrefixMessage)
	KeyMessagesPrivate = NewCacheKey(KeyPrefixMessagesPrivate)
	KeyMessagesGroup   = NewCacheKey(KeyPrefixMessagesGroup)
	KeyConversations   = NewCacheKey(KeyPrefixConversations)
	KeyFriends         = NewCacheKey(KeyPrefixFriends)
	KeyFriendRequests  = NewCacheKey(KeyPrefixFriendRequests)
	KeyRateLimit       = NewCacheKey(KeyPrefixRateLimit)
)

// CacheError 缓存错误类型
type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache error: operation=%s, key=%s, error=%v", e.Operation, e.Key, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// IsRedisNil 检查是否为 Redis Nil 错误
func IsRedisNil(err error) bool {
	return err == redis.Nil
}

// IsCacheError 检查是否为缓存错误
func IsCacheError(err error) bool {
	_, ok := err.(*CacheError)
	return ok
}

// WrapCacheError 包装缓存错误
func WrapCacheError(operation, key string, err error) error {
	if err == nil {
		return nil
	}
	return &CacheError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}
