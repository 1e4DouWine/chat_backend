package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"chat_backend/internal/config"
)

var (
	// db GORM数据库实例
	db *gorm.DB
	// redisClient Redis客户端
	redisClient *redis.Client
	// once 确保只初始化一次
	once sync.Once
	// initErr 初始化错误
	initErr error
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	once.Do(func() {
		// 初始化PostgreSQL连接
		if err := initPostgreSQLWithGORM(cfg.Database); err != nil {
			initErr = fmt.Errorf("初始化PostgreSQL失败: %w", err)
			return
		}

		// 初始化Redis连接
		if err := initRedis(cfg.Redis); err != nil {
			initErr = fmt.Errorf("初始化Redis失败: %w", err)
			return
		}
	})
	return initErr
}

// initPostgreSQLWithGORM 使用GORM初始化PostgreSQL连接
func initPostgreSQLWithGORM(cfg config.DatabaseConfig) error {
	// 构建DSN连接字符串
	dsn := cfg.GetDSN()

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 创建数据库连接
	database, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("创建GORM连接失败: %w", err)
	}

	// 获取底层SQL数据库对象以进行连接测试
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("获取SQL数据库对象失败: %w", err)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		err := sqlDB.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("连接测试失败: %w", err)
	}

	// 设置连接池参数（从配置文件读取，如果未配置则使用默认值）
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25 // 默认最大打开连接数
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 5 // 默认最大空闲连接数
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	db = database
	return nil
}

// initRedis 初始化Redis连接
func initRedis(cfg config.RedisConfig) error {
	ctx := context.Background()

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		err := client.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("redis连接测试失败: %w", err)
	}

	redisClient = client
	return nil
}

// GetDB 获取GORM数据库实例
func GetDB() *gorm.DB {
	if db == nil {
		panic("数据库连接未初始化")
	}
	return db
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	if redisClient == nil {
		panic("Redis连接未初始化")
	}
	return redisClient
}

// Close 关闭所有数据库连接
func Close() error {
	var errs []error

	// 关闭PostgreSQL连接
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("获取SQL数据库对象失败: %w", err))
		} else {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("关闭PostgreSQL连接失败: %w", err))
			}
		}
	}

	// 关闭Redis连接
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Redis连接失败: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}
	return nil
}

// HealthCheck 数据库健康检查
func HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)

	// PostgreSQL健康检查
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			status["postgresql"] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			if err := sqlDB.Ping(); err != nil {
				status["postgresql"] = fmt.Sprintf("unhealthy: %v", err)
			} else {
				status["postgresql"] = "healthy"
			}
		}
	} else {
		status["postgresql"] = "not initialized"
	}

	// Redis健康检查
	if redisClient != nil {
		if err := redisClient.Ping(ctx).Err(); err != nil {
			status["redis"] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			status["redis"] = "healthy"
		}
	} else {
		status["redis"] = "not initialized"
	}

	return status
}
