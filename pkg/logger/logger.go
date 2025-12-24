package logger

import (
	"chat_backend/pkg/env"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志记录器实例
	Logger *zap.SugaredLogger
)

// Init 初始化全局日志记录器
// 根据环境变量配置不同的日志输出方式：
// - 本地开发环境：输出到控制台和文件，控制台带颜色，文件不带颜色
// - 测试/生产环境：只输出到文件，不带颜色
// - 所有环境都将日志写入app.log文件
func Init() error {
	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 设置日志文件路径
	logFile := filepath.Join(logDir, "app.log")

	// 创建日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建核心配置
	core := zapcore.NewCore(
		getFileEncoder(), // 文件编码器，不带颜色
		zapcore.AddSync(file),
		zap.InfoLevel,
	)

	var logger *zap.Logger

	if env.IsLocalDev() {
		// 本地开发环境：同时输出到控制台和文件
		consoleCore := zapcore.NewCore(
			getConsoleEncoder(), // 控制台编码器，带颜色
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
		// 使用多输出核心，同时输出到控制台和文件
		teeCore := zapcore.NewTee(core, consoleCore)
		logger = zap.New(teeCore, zap.AddCaller())
	} else {
		// 测试/生产环境：只输出到文件
		logger = zap.New(core, zap.AddCaller())
	}

	Logger = logger.Sugar()

	// 测试日志记录器
	Logger.Info("日志记录器初始化成功")

	return nil
}

// getFileEncoder 返回用于文件输出的编码器，不带颜色
func getFileEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 不带颜色的级别编码器
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getConsoleEncoder 返回用于控制台输出的编码器，带颜色
func getConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.StacktraceKey = ""
	encoderConfig.CallerKey = ""
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 带颜色的级别编码器
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// Sync 刷新所有缓冲的日志条目
// 此方法应在应用程序退出前调用，以确保
// 所有缓冲的日志都被写入
// 如果同步失败则返回错误，否则返回 nil
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// GetLogger 返回全局日志记录器实例
// 如果日志记录器未初始化，它将创建一个回退的生产环境日志记录器
// 返回全局 SugaredLogger 实例
func GetLogger() *zap.SugaredLogger {
	if Logger == nil {
		// 如果未初始化，回退到简单的日志记录器
		baseLogger, _ := zap.NewProduction()
		Logger = baseLogger.Sugar()
	}
	return Logger
}
