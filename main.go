package main

import (
	"chat_backend/internal/config"
	"chat_backend/internal/database"
	"chat_backend/internal/middleware"
	"chat_backend/internal/router"
	"chat_backend/pkg/logger"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	// 解析命令行参数
	var migrateFlag = flag.Bool("migrate", false, "执行数据库迁移")
	var resetFlag = flag.Bool("reset-db", false, "重置数据库（删除并重新创建所有表）")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化全局日志记录器
	if err := logger.Init(); err != nil {
		panic(fmt.Sprintf("初始化日志记录器失败: %v", err))
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.GetLogger().Errorw("同步日志记录器失败", "error", err)
		}
	}()

	// 初始化数据库和redis连接
	if err := database.InitDB(cfg); err != nil {
		logger.GetLogger().Fatalw("初始化数据库失败", "error", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.GetLogger().Errorw("关闭数据库连接失败", "error", err)
		}
	}()

	// 处理数据库迁移命令
	if *resetFlag {
		// 重置数据库
		logger.GetLogger().Infow("正在重置数据库...")
		if err := database.ResetDatabase(); err != nil {
			logger.GetLogger().Fatalw("重置数据库失败", "error", err)
		}
		logger.GetLogger().Infow("数据库重置成功")
		os.Exit(0)
	}

	if *migrateFlag {
		// 执行数据库迁移
		logger.GetLogger().Infow("正在执行数据库迁移...")
		if err := database.Migrate(); err != nil {
			logger.GetLogger().Fatalw("数据库迁移失败", "error", err)
		}
		logger.GetLogger().Infow("数据库迁移成功")
		os.Exit(0)
	}

	// 数据库健康检查
	ctx := context.Background()
	status := database.HealthCheck(ctx)
	for service, state := range status {
		if state != "healthy" {
			logger.GetLogger().Warnw("数据库服务未健康", "service", service, "status", state)
		} else {
			logger.GetLogger().Infow("数据库服务健康", "service", service)
		}
	}

	// 初始化JWT配置
	middleware.InitJWT(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessExpiry)*time.Hour,
		time.Duration(cfg.JWT.RefreshExpiry)*time.Hour,
	)

	startServer(cfg)
}

func startServer(cfg *config.Config) {
	e := echo.New()

	// 添加中间件
	e.Use(middleware.CORSMiddleware()) // 添加CORS中间件
	e.Use(middleware.RecoverMiddleware())
	e.Use(middleware.LoggerMiddleware())

	// 初始化路由
	router.InitRouter(e)

	// 示例路由
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "欢迎使用聊天后端API",
		})
	})

	// 健康检查路由
	e.GET("/health", func(c echo.Context) error {
		ctx := c.Request().Context()
		status := database.HealthCheck(ctx)

		// 检查所有服务是否健康
		allHealthy := true
		for _, state := range status {
			if state != "healthy" {
				allHealthy = false
				break
			}
		}

		if allHealthy {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":   "healthy",
				"services": status,
			})
		} else {
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status":   "unhealthy",
				"services": status,
			})
		}
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.GetLogger().Infow("正在启动服务器", "address", addr)

	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.GetLogger().Fatalw("启动服务器失败", "error", err)
	}
}
