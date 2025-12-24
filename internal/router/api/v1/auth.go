package v1

import (
	"chat_backend/internal/database"
	"chat_backend/internal/dto"
	"chat_backend/internal/response"
	"chat_backend/internal/service"
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

// Register 用户注册
func Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 1000, "invalid request format")
	}

	// 基本验证
	if strings.TrimSpace(req.Username) == "" {
		return response.Error(c, 1001, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return response.Error(c, 1001, "password is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameAlreadyExists) {
			return response.Error(c, 5001, "username already exists")
		}
		return response.Error(c, 5000, "failed to register user")
	}

	return response.Success(c, authResp)
}

// Login 用户登录
func Login(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 1000, "invalid request format")
	}

	// 基本验证
	if strings.TrimSpace(req.Username) == "" {
		return response.Error(c, 1001, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return response.Error(c, 1001, "password is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.Login(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUsernameOrPassword) {
			return response.Error(c, 2002, "invalid username or password")
		}
		return response.Error(c, 5000, "failed to login")
	}

	return response.Success(c, authResp)
}

// RefreshToken 刷新访问令牌
func RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 1000, "invalid request format")
	}

	// 基本验证
	if strings.TrimSpace(req.RefreshToken) == "" {
		return response.Error(c, 1001, "refresh_token is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.RefreshToken(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid refresh token") ||
			strings.Contains(err.Error(), "user not found") {
			return response.Error(c, 2001, "invalid refresh token")
		}
		return response.Error(c, 5000, "failed to refresh token")
	}

	return response.Success(c, authResp)
}
