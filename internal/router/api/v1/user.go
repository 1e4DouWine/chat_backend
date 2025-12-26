package v1

import (
	"chat_backend/internal/database"
	"chat_backend/internal/dto"
	"chat_backend/internal/errors"
	"chat_backend/internal/global"
	"chat_backend/internal/response"
	"chat_backend/internal/service"

	"github.com/labstack/echo/v4"
)

const (
	getFriendListQueryParam = "status"

	processFriendRequestParam = "action"
)

// GetMe 获取当前用户的信息
func GetMe(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get(global.JwtKeyUserID).(string)
	username := c.Get(global.JwtKeyUserName).(string)

	if userID == "" {
		return response.Error(c, errors.ErrCodeUserIDRequired, errors.GetMessage(errors.ErrCodeUserIDRequired))
	}

	userService := service.NewUserService(database.GetDB())

	userInfoResponse, err := userService.GetMe(ctx, userID, username)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToGetUserInfo, errors.GetMessage(errors.ErrCodeFailedToGetUserInfo))
	}

	return response.Success(c, userInfoResponse)
}

// AddFriend 添加朋友
func AddFriend(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.AddFriendRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}
	if req.Username == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "username is required")
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())
	friendID, err := userService.GetUserIDByUsername(ctx, req.Username)
	if err != nil {
		return response.Error(c, errors.ErrCodeUserNotFound, errors.GetMessage(errors.ErrCodeUserNotFound))
	}

	resp, err := userService.AddFriend(ctx, userID, friendID)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToAddFriend, err.Error())
	}
	return response.Success(c, resp)
}

// GetFriendList 获取好友列表
func GetFriendList(c echo.Context) error {
	ctx := c.Request().Context()
	status := c.QueryParam(getFriendListQueryParam)
	if status == "" {
		status = service.FriendStatusAccepted
	} else if status != service.FriendStatusPending && status != service.FriendStatusAccepted {
		return response.Error(c, errors.ErrCodeInvalidStatus, errors.GetMessage(errors.ErrCodeInvalidStatus))
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())

	friendList, err := userService.GetFriendList(ctx, userID, status)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToGetFriendList, errors.GetMessage(errors.ErrCodeFailedToGetFriendList))
	}
	return response.Success(c, friendList)
}

// ProcessFriendRequest 处理好友申请
func ProcessFriendRequest(c echo.Context) error {
	ctx := c.Request().Context()
	friendID := c.Param("id")
	action := c.QueryParam(processFriendRequestParam)
	if action != service.ActionParamAccept && action != service.ActionParamReject {
		return response.Error(c, errors.ErrCodeInvalidAction, errors.GetMessage(errors.ErrCodeInvalidAction))
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())
	resp, err := userService.ProcessFriendRequest(ctx, userID, friendID, action)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToProcessRequest, err.Error())
	}
	return response.Success(c, resp)
}

// DeleteFriend 删除好友
func DeleteFriend(c echo.Context) error {
	ctx := c.Request().Context()
	friendID := c.Param("id")
	userID := c.Get(global.JwtKeyUserID).(string)
	userService := service.NewUserService(database.GetDB())

	err := userService.DeleteFriend(ctx, userID, friendID)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToDeleteFriend, err.Error())
	}
	return response.Success(c, nil)
}
