package v1

import (
	"chat_backend/internal/database"
	"chat_backend/internal/dto"
	"chat_backend/internal/errors"
	"chat_backend/internal/response"
	"chat_backend/internal/service"
	"strings"

	"github.com/labstack/echo/v4"
)

// Register 用户注册
func Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	// 基本验证
	if strings.TrimSpace(req.Username) == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "password is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.Register(ctx, req)
	if err != nil {
		if err.Error() == service.ErrUsernameAlreadyExists.Error() {
			return response.Error(c, errors.ErrCodeUsernameAlreadyExists, errors.GetMessage(errors.ErrCodeUsernameAlreadyExists))
		}
		return response.Error(c, errors.ErrCodeFailedToRegister, errors.GetMessage(errors.ErrCodeFailedToRegister))
	}

	return response.Success(c, authResp)
}

// Login 用户登录
func Login(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	// 基本验证
	if strings.TrimSpace(req.Username) == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "password is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.Login(ctx, req)
	if err != nil {
		if err.Error() == service.ErrInvalidUsernameOrPassword.Error() {
			return response.Error(c, errors.ErrCodeInvalidCredentials, errors.GetMessage(errors.ErrCodeInvalidCredentials))
		}
		return response.Error(c, errors.ErrCodeFailedToLogin, errors.GetMessage(errors.ErrCodeFailedToLogin))
	}

	return response.Success(c, authResp)
}

// RefreshToken 刷新访问令牌
func RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	// 基本验证
	if strings.TrimSpace(req.RefreshToken) == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "refresh_token is required")
	}

	// 获取服务实例
	authService := service.NewAuthService(database.GetDB())

	// 调用服务层处理业务逻辑
	authResp, err := authService.RefreshToken(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid refresh token") ||
			strings.Contains(err.Error(), "user not found") {
			return response.Error(c, errors.ErrCodeInvalidRefreshToken, errors.GetMessage(errors.ErrCodeInvalidRefreshToken))
		}
		return response.Error(c, errors.ErrCodeFailedToRefreshToken, errors.GetMessage(errors.ErrCodeFailedToRefreshToken))
	}

	return response.Success(c, authResp)
}
