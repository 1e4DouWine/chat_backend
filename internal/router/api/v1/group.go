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

// CreateGroup 创建群组
func CreateGroup(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}
	if req.Name == "" {
		return response.Error(c, errors.ErrCodeGroupNameEmpty, errors.GetMessage(errors.ErrCodeGroupNameEmpty))
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	group, err := groupService.CreateGroup(ctx, userID, req.Name)
	if err != nil {
		return response.Error(c, errors.ErrCodeFailedToCreateGroup, err.Error())
	}

	return response.Success(c, group)
}

// GetGroupList 获取群组列表
func GetGroupList(c echo.Context) error {
	ctx := c.Request().Context()
	role := c.QueryParam("role")
	if role != "" {
		if role != service.RoleOwner && role != service.RoleMember {
			return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
		}
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	groupList, err := groupService.GetGroupList(ctx, userID, role)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, groupList)
}

// GetGroupDetail 获取群组详情
func GetGroupDetail(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("id")
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group id is required")
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	groupDetail, err := groupService.GetGroupDetail(ctx, userID, groupID)
	if err != nil {
		if err.Error() == "you are not in this group" {
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		}
		return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
	}

	return response.Success(c, groupDetail)
}

// JoinGroup 加入群组
func JoinGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("id")
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group id is required")
	}
	var req dto.JoinGroupRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	result, err := groupService.JoinGroup(ctx, userID, groupID, req.InviteCode)
	if err != nil {
		switch err.Error() {
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "already in group":
			return response.Error(c, errors.ErrCodeAlreadyInGroup, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToJoinGroup, err.Error())
		}
	}

	return response.Success(c, result)
}

// LeaveGroup 退出群组
func LeaveGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("id")

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group id is required")
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.LeaveGroup(ctx, userID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "not in group":
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case "cannot leave as owner":
			return response.Error(c, errors.ErrCodeCannotLeaveAsOwner, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToLeaveGroup, err.Error())
		}
	}

	return response.Success(c, nil)
}

// DisbandGroup 解散群组
func DisbandGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("id")

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group id is required")
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.DisbandGroup(ctx, userID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "permission denied":
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToDisbandGroup, err.Error())
		}
	}

	return response.Success(c, nil)
}

// TransferGroup 转让群组
func TransferGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("id")

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group id is required")
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	var req dto.TransferGroupRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	if req.NewOwnerID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "new_owner_id is required")
	}

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.TransferGroup(ctx, userID, groupID, req.NewOwnerID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "permission denied":
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case "target user not in group":
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToTransferGroup, err.Error())
		}
	}

	return response.Success(c, nil)
}

// RemoveMember 移除群组成员
func RemoveMember(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("group_id")
	targetUserID := c.Param("user_id")

	if groupID == "" || targetUserID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "group_id and user_id are required")
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.RemoveMember(ctx, userID, groupID, targetUserID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "permission denied":
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case "target user not in group":
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		case "cannot remove yourself":
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		case "cannot remove owner":
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToRemoveMember, err.Error())
		}
	}

	return response.Success(c, nil)
}

// JoinGroupByCode 通过邀请码加入群组
func JoinGroupByCode(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.JoinGroupByCodeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}
	if req.InviteCode == "" {
		return response.Error(c, errors.ErrCodeInviteCodeRequired, errors.GetMessage(errors.ErrCodeInviteCodeRequired))
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	result, err := groupService.JoinGroupByCode(ctx, userID, req.InviteCode)
	if err != nil {
		switch err.Error() {
		case "invalid invite code":
			return response.Error(c, errors.ErrCodeInvalidInviteCode, err.Error())
		case "group not found":
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case "already in group":
			return response.Error(c, errors.ErrCodeAlreadyInGroup, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToJoinGroup, err.Error())
		}
	}

	return response.Success(c, result)
}

// SearchGroup 搜索群组
func SearchGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupName := c.QueryParam("name")
	if groupName == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, "name is required")
	}

	groupService := service.NewGroupService(database.GetDB())
	groups, err := groupService.SearchGroup(ctx, groupName)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, groups)
}
