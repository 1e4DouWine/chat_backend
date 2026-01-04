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
	role := c.QueryParam(QueryParamRole)
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
	groupID := c.Param(ParamID)
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	groupDetail, err := groupService.GetGroupDetail(ctx, userID, groupID)
	if err != nil {
		if err.Error() == ErrorMessageYouAreNotInThisGroup {
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		}
		return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
	}

	return response.Success(c, groupDetail)
}

// JoinGroup 加入群组
func JoinGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param(ParamID)
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
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
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessageAlreadyInGroup:
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
	groupID := c.Param(ParamID)

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.LeaveGroup(ctx, userID, groupID)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessageNotInGroup:
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case ErrorMessageCannotLeaveAsOwner:
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
	groupID := c.Param(ParamID)

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.DisbandGroup(ctx, userID, groupID)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessagePermissionDenied:
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
	groupID := c.Param(ParamID)

	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	var req dto.TransferGroupRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	if req.NewOwnerID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageNewOwnerIDRequired)
	}

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.TransferGroup(ctx, userID, groupID, req.NewOwnerID)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessagePermissionDenied:
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case ErrorMessageTargetUserNotInGroup:
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
	groupID := c.Param(ParamGroupID)
	targetUserID := c.Param(ParamUserID)

	if groupID == "" || targetUserID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDAndUserIDRequired)
	}
	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.RemoveMember(ctx, userID, groupID, targetUserID)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessagePermissionDenied:
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case ErrorMessageTargetUserNotInGroup:
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		case ErrorMessageCannotRemoveYourself:
			return response.Error(c, errors.ErrCodeInvalidRequest, err.Error())
		case ErrorMessageCannotRemoveOwner:
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
		case ErrorMessageInvalidInviteCode:
			return response.Error(c, errors.ErrCodeInvalidInviteCode, err.Error())
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessageAlreadyInGroup:
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
	groupName := c.QueryParam(QueryParamName)
	if groupName == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageNameRequired)
	}

	groupService := service.NewGroupService(database.GetDB())
	groups, err := groupService.SearchGroup(ctx, groupName)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, groups)
}

// RequestJoinGroup 申请加入群组
func RequestJoinGroup(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param(ParamID)
	if groupID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDRequired)
	}

	var req dto.RequestJoinGroupRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	result, err := groupService.RequestJoinGroup(ctx, userID, groupID, req.Message)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessageAlreadyInGroup:
			return response.Error(c, errors.ErrCodeAlreadyInGroup, err.Error())
		case ErrorMessagePendingRequestAlreadyExists:
			return response.Error(c, errors.ErrCodeAlreadyRequested, err.Error())
		case ErrorMessageCannotRequestWithinCooldown:
			return response.Error(c, errors.ErrCodeCannotRequestWithinCooldown, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToRequestJoinGroup, err.Error())
		}
	}

	return response.Success(c, result)
}

// GetPendingJoinRequests 获取待审核的入群请求
func GetPendingJoinRequests(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	requests, err := groupService.GetPendingJoinRequests(ctx, userID)
	if err != nil {
		return response.Error(c, errors.ErrCodeInternalError, err.Error())
	}

	return response.Success(c, requests)
}

// ApproveJoinRequest 审批入群请求
func ApproveJoinRequest(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param(ParamID)
	senderID := c.Param(ParamUserID)

	if groupID == "" || senderID == "" {
		return response.Error(c, errors.ErrCodeRequiredFieldMissing, ErrorMessageGroupIDAndUserIDRequired2)
	}

	var req dto.ApproveJoinRequestRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.ErrCodeInvalidRequest, errors.GetMessage(errors.ErrCodeInvalidRequest))
	}

	if req.Action != "approve" && req.Action != "reject" {
		return response.Error(c, errors.ErrCodeInvalidAction, ErrorMessageActionMustBeApproveOrReject)
	}

	userID := c.Get(global.JwtKeyUserID).(string)

	groupService := service.NewGroupService(database.GetDB())
	err := groupService.ApproveJoinRequest(ctx, userID, groupID, senderID, req.Action)
	if err != nil {
		switch err.Error() {
		case ErrorMessageGroupNotFound:
			return response.Error(c, errors.ErrCodeGroupNotFound, err.Error())
		case ErrorMessagePermissionDenied:
			return response.Error(c, errors.ErrCodePermissionDenied, err.Error())
		case ErrorMessageJoinRequestNotFound:
			return response.Error(c, errors.ErrCodeJoinRequestNotFound, err.Error())
		case ErrorMessageInvalidAction:
			return response.Error(c, errors.ErrCodeInvalidAction, err.Error())
		case ErrorMessageAlreadyInGroup:
			return response.Error(c, errors.ErrCodeAlreadyInGroup, err.Error())
		default:
			return response.Error(c, errors.ErrCodeFailedToApproveJoinRequest, err.Error())
		}
	}

	return response.Success(c, nil)
}
