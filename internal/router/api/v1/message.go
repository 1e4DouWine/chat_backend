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

	targetUserID := c.QueryParam("target_user_id")
	if targetUserID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "target_user_id is required")
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())
	isFriend, err := userService.IsFriend(ctx, userID, targetUserID)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}
	if !isFriend {
		return response.Error(c, errors.ErrCodePermissionDenied, "You are not in a friend relationship")
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

// GetConversationList 获取会话列表
func GetConversationList(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(global.JwtKeyUserID).(string)

	messageService := service.NewMessageService(database.GetDB())
	result, err := messageService.GetConversationList(ctx, userID)
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

	userID := c.Get(global.JwtKeyUserID).(string)
	groupService := service.NewGroupService(database.GetDB())
	isGroupMember, err := groupService.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}
	if !isGroupMember {
		return response.Error(c, errors.ErrCodePermissionDenied, "You are not a member of this group")
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
