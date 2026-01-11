package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"chat_backend/internal/cache"
)

const (
	// DefaultRateLimitWindow 默认限流时间窗口（1分钟）
	DefaultRateLimitWindow = 1 * time.Minute
	// DefaultRateLimitCount 默认限流请求数量（60次/分钟）
	DefaultRateLimitCount = 60
	// RateLimitHeaderLimit 限流响应头
	RateLimitHeaderLimit = "X-RateLimit-Limit"
	// RateLimitHeaderRemaining 剩余请求数响应头
	RateLimitHeaderRemaining = "X-RateLimit-Remaining"
	// RateLimitHeaderReset 重置时间响应头
	RateLimitHeaderReset = "X-RateLimit-Reset"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Window time.Duration // 时间窗口
	Count  int64         // 请求次数限制
	Key    string        // 限流键前缀
}

// RateLimiter 限流器
type RateLimiter struct {
	redis *redis.Client
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
	}
}

// RateLimitMiddleware 创建限流中间件
func RateLimitMiddleware(limiter *RateLimiter, config RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// 生成限流键
			key := generateRateLimitKey(c, config.Key)

			// 检查限流
			allowed, remaining, resetAt, err := limiter.Allow(ctx, key, config.Window, config.Count)
			if err != nil {
				// 限流检查失败，记录错误但允许请求通过
				c.Logger().Errorf("Rate limit check failed: %v", err)
				return next(c)
			}

			// 设置响应头
			c.Response().Header().Set(RateLimitHeaderLimit, strconv.FormatInt(config.Count, 10))
			c.Response().Header().Set(RateLimitHeaderRemaining, strconv.FormatInt(remaining, 10))
			c.Response().Header().Set(RateLimitHeaderReset, strconv.FormatInt(resetAt, 10))

			// 如果超过限制，返回 429
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":       "Too many requests",
					"retry_after": resetAt - time.Now().Unix(),
				})
			}

			return next(c)
		}
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ctx context.Context, key string, window time.Duration, limit int64) (bool, int64, int64, error) {
	now := time.Now()
	currentTime := now.Unix()

	// 使用 Lua 脚本保证原子性
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local current_time = tonumber(ARGV[3])

		-- 删除时间窗口外的记录
		redis.call('ZREMRANGEBYSCORE', key, '-inf', current_time - window)

		-- 获取当前请求数
		local current = redis.call('ZCARD', key)

		-- 检查是否超过限制
		if current < limit then
			-- 添加当前请求
			redis.call('ZADD', key, current_time, current_time)
			-- 设置过期时间
			redis.call('EXPIRE', key, window)
			return {1, limit - current - 1, current_time + window}
		else
			-- 获取最早的请求时间
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset_time = oldest[2] or current_time + window
			return {0, 0, reset_time}
		end
	`

	result, err := rl.redis.Eval(ctx, luaScript, []string{key}, window.Seconds(), limit, currentTime).Result()
	if err != nil {
		return false, 0, 0, err
	}

	// 解析 Lua 脚本返回结果
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return false, 0, 0, fmt.Errorf("invalid result from lua script")
	}

	allowed := resultSlice[0].(int64) == 1
	remaining := resultSlice[1].(int64)
	resetAt := resultSlice[2].(int64)

	return allowed, remaining, resetAt, nil
}

// generateRateLimitKey 生成限流键
func generateRateLimitKey(c echo.Context, prefix string) string {
	// 优先使用用户ID
	if userID := c.Get("user_id"); userID != nil {
		return fmt.Sprintf("%s:user:%s", prefix, userID)
	}

	// 其次使用IP地址
	if ip := c.RealIP(); ip != "" {
		return fmt.Sprintf("%s:ip:%s", prefix, ip)
	}

	// 最后使用默认键
	return fmt.Sprintf("%s:default", prefix)
}

// SimpleRateLimitMiddleware 简单限流中间件（使用 INCR + EXPIRE）
func SimpleRateLimitMiddleware(limiter *RateLimiter, config RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// 生成限流键
			key := generateRateLimitKey(c, config.Key)

			// 检查限流
			allowed, remaining, resetAt, err := limiter.SimpleAllow(ctx, key, config.Window, config.Count)
			if err != nil {
				// 限流检查失败，记录错误但允许请求通过
				c.Logger().Errorf("Rate limit check failed: %v", err)
				return next(c)
			}

			// 设置响应头
			c.Response().Header().Set(RateLimitHeaderLimit, strconv.FormatInt(config.Count, 10))
			c.Response().Header().Set(RateLimitHeaderRemaining, strconv.FormatInt(remaining, 10))
			c.Response().Header().Set(RateLimitHeaderReset, strconv.FormatInt(resetAt, 10))

			// 如果超过限制，返回 429
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":       "Too many requests",
					"retry_after": resetAt - time.Now().Unix(),
				})
			}

			return next(c)
		}
	}
}

// SimpleAllow 简单限流检查（使用 INCR + EXPIRE）
func (rl *RateLimiter) SimpleAllow(ctx context.Context, key string, window time.Duration, limit int64) (bool, int64, int64, error) {
	now := time.Now()
	currentTime := now.Unix()

	// 使用 Pipeline 减少网络往返
	pipe := rl.redis.Pipeline()

	// 递增计数器
	incrCmd := pipe.Incr(ctx, key)

	// 设置过期时间（仅在第一次设置）
	pipe.Expire(ctx, key, window)

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	// 获取当前计数
	count := incrCmd.Val()

	// 计算剩余次数
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	// 计算重置时间
	resetAt := currentTime + int64(window.Seconds())

	// 检查是否超过限制
	if count > limit {
		return false, remaining, resetAt, nil
	}

	return true, remaining, resetAt, nil
}

// GetRateLimitKey 生成限流键（供外部使用）
func GetRateLimitKey(prefix, userID, ip string) string {
	if userID != "" {
		return fmt.Sprintf("%s:user:%s", prefix, userID)
	}
	if ip != "" {
		return fmt.Sprintf("%s:ip:%s", prefix, ip)
	}
	return fmt.Sprintf("%s:default", prefix)
}

// Predefined rate limit configurations
var (
	// AuthRateLimit 认证接口限流（10次/分钟）
	AuthRateLimit = RateLimitConfig{
		Window: 1 * time.Minute,
		Count:  10,
		Key:    cache.KeyRateLimit.Build("auth"),
	}

	// MessageRateLimit 消息发送限流（60次/分钟）
	MessageRateLimit = RateLimitConfig{
		Window: 1 * time.Minute,
		Count:  60,
		Key:    cache.KeyRateLimit.Build("message"),
	}

	// GeneralRateLimit 通用接口限流（60次/分钟）
	GeneralRateLimit = RateLimitConfig{
		Window: 1 * time.Minute,
		Count:  60,
		Key:    cache.KeyRateLimit.Build("general"),
	}

	// UploadRateLimit 上传接口限流（10次/分钟）
	UploadRateLimit = RateLimitConfig{
		Window: 1 * time.Minute,
		Count:  10,
		Key:    cache.KeyRateLimit.Build("upload"),
	}
)
