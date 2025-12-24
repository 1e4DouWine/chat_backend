package v1

import (
	"chat_backend/internal/database"
	"chat_backend/internal/dto"
	"chat_backend/internal/global"
	"chat_backend/internal/response"
	"chat_backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetMe 获取当前用户的信息
func GetMe(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get(global.JwtKeyUserID).(string)
	username := c.Get(global.JwtKeyUserName).(string)

	if userID == "" {
		return response.Error(c, http.StatusBadRequest, "userID is required")
	}

	userService := service.NewUserService(database.GetDB())

	userInfoResponse, err := userService.GetMe(ctx, userID, username)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}

	return response.Success(c, userInfoResponse)
}

// AddFriend 添加朋友
func AddFriend(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.AddFriendRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 1000, "invalid request format")
	}
	if req.Username == "" {
		return response.Error(c, 1000, "username is required")
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())
	friendID, err := userService.GetUserIDByUsername(ctx, req.Username)
	if err != nil {
		return response.Error(c, 1000, "no user found")
	}

	resp, err := userService.AddFriend(ctx, userID, friendID)
	if err != nil {
		return response.Error(c, 1000, "add friend error")
	}
	return response.Success(c, resp)
}
