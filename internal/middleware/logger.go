package middleware

import (
	"chat_backend/pkg/logger"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// LoggerMiddleware 返回一个记录HTTP请求的中间件
func LoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// 处理请求
			err := next(c)

			// 计算延迟
			latency := time.Since(start)

			// 获取请求详情
			req := c.Request()
			res := c.Response()

			// 获取客户端IP
			clientIP := c.RealIP()

			// 获取实际的状态码，包括错误情况
			status := res.Status
			if err != nil {
				// 如果有错误，尝试从HTTP错误中获取状态码
				var he *echo.HTTPError
				if errors.As(err, &he) {
					status = he.Code
				}
			}

			// 记录请求详情
			fields := []interface{}{
				"method", req.Method,
				"uri", req.RequestURI,
				"host", req.Host,
				"remote_ip", clientIP,
				"user_agent", req.UserAgent(),
				"status", status,
				"latency", latency,
				"latency_human", latency.String(),
			}

			// 添加路径参数（如果有）
			if len(c.ParamNames()) > 0 {
				params := make(map[string]string)
				for i, name := range c.ParamNames() {
					params[name] = c.ParamValues()[i]
				}
				fields = append(fields, "params", params)
			}

			// 添加查询参数（如果有）
			if len(c.QueryParams()) > 0 {
				fields = append(fields, "query", c.QueryParams())
			}

			// 根据状态码记录日志
			switch {
			case status >= 500:
				logger.GetLogger().Errorw("服务器错误", fields...)
			case status >= 400:
				logger.GetLogger().Warnw("客户端错误", fields...)
			case status >= 300:
				logger.GetLogger().Infow("重定向", fields...)
			default:
				logger.GetLogger().Infow("请求完成", fields...)
			}

			return err
		}
	}
}

// RecoverMiddleware 返回一个从panic中恢复并记录错误的中间件
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = echo.NewHTTPError(http.StatusInternalServerError, r)
					}

					logger.GetLogger().Errorw("从panic中恢复",
						"method", c.Request().Method,
						"uri", c.Request().RequestURI,
						"remote_ip", c.RealIP(),
						"error", err,
					)

					c.Error(err)
				}
			}()

			return next(c)
		}
	}
}
