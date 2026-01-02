package v1

import (
	"chat_backend/internal/database"
	"chat_backend/internal/errors"
	"chat_backend/internal/global"
	"chat_backend/internal/response"
	"chat_backend/internal/service"

	"strconv"

	"github.com/labstack/echo/v4"
)

// GetPrivateMessages 获取私聊消息记录
func GetPrivateMessages(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(global.JwtKeyUserID).(string)

	targetUserID := c.QueryParam("target_user_id")
	if targetUserID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "target_user_id is required")
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	cursor := c.QueryParam("cursor")

	messageService := service.NewMessageService(database.GetDB())
	result, err := messageService.GetPrivateMessages(ctx, userID, targetUserID, limit, cursor)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, result)
}

// GetGroupMessages 获取群聊消息记录
func GetGroupMessages(c echo.Context) error {
	ctx := c.Request().Context()

	groupID := c.Param("id")
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group_id is required")
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	cursor := c.QueryParam("cursor")

	messageService := service.NewMessageService(database.GetDB())
	result, err := messageService.GetGroupMessages(ctx, groupID, limit, cursor)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, result)
}
