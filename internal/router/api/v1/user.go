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
		return response.Error(c, 1000, err.Error())
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
		return response.Error(c, 1000, "invalid status parameter")
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())

	friendList, err := userService.GetFriendList(ctx, userID, status)
	if err != nil {
		return response.Error(c, 1000, err.Error())
	}
	return response.Success(c, friendList)
}

// ProcessFriendRequest 处理好友申请
func ProcessFriendRequest(c echo.Context) error {
	ctx := c.Request().Context()
	friendID := c.Param("id")
	action := c.QueryParam(processFriendRequestParam)
	if action != service.ActionParamAccept && action != service.ActionParamReject {
		return response.Error(c, 1000, "invalid action parameter")
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	userService := service.NewUserService(database.GetDB())
	resp, err := userService.ProcessFriendRequest(ctx, userID, friendID, action)
	if err != nil {
		return response.Error(c, 1000, err.Error())
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
		return response.Error(c, 1000, err.Error())
	}
	return response.Success(c, nil)
}
