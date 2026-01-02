package errors

const (
	Success = 0
)

const (
	ErrCodeInvalidRequest       = 1000
	ErrCodeRequiredFieldMissing = 1001
	ErrCodeInvalidStatus        = 1002
	ErrCodeInvalidAction        = 1003
	ErrCodeUserNotFound         = 1004
)

const (
	ErrCodeInvalidRefreshToken = 2001
	ErrCodeInvalidCredentials  = 2002
)

const (
	ErrCodeUserIDRequired = 4000
)

const (
	ErrCodeInternalError          = 5000
	ErrCodeUsernameAlreadyExists  = 5001
	ErrCodeFailedToRegister       = 5002
	ErrCodeFailedToLogin          = 5003
	ErrCodeFailedToRefreshToken   = 5004
	ErrCodeFailedToGetUserInfo    = 5005
	ErrCodeFailedToAddFriend      = 5006
	ErrCodeFailedToGetFriendList  = 5007
	ErrCodeFailedToProcessRequest = 5008
	ErrCodeFailedToDeleteFriend   = 5009
	ErrCodeGroupNameEmpty         = 5010
	ErrCodeFailedToCreateGroup    = 5011
	ErrCodeGroupNotFound          = 5012
	ErrCodeGroupAlreadyExists     = 5013
	ErrCodePermissionDenied       = 5014
	ErrCodeInvalidGroupRole       = 5015
	ErrCodeFailedToJoinGroup      = 5016
	ErrCodeFailedToLeaveGroup     = 5017
	ErrCodeFailedToDisbandGroup   = 5018
	ErrCodeFailedToTransferGroup  = 5019
	ErrCodeFailedToRemoveMember   = 5020
	ErrCodeInviteCodeRequired     = 5021
	ErrCodeInvalidInviteCode      = 5022
	ErrCodeAlreadyInGroup         = 5023
	ErrCodeCannotLeaveAsOwner     = 5024
	ErrCodeFailedToRequestJoinGroup      = 5025
	ErrCodeJoinRequestNotFound           = 5026
	ErrCodeFailedToApproveJoinRequest    = 5027
	ErrCodeCannotRequestWithinCooldown  = 5028
	ErrCodeAlreadyRequested             = 5029
)

var (
	ErrMessages = map[int]string{
		ErrCodeInvalidRequest:         "invalid request format",
		ErrCodeRequiredFieldMissing:   "required field is missing",
		ErrCodeInvalidStatus:          "invalid status parameter",
		ErrCodeInvalidAction:          "invalid action parameter",
		ErrCodeUserNotFound:           "user not found",
		ErrCodeInvalidRefreshToken:    "invalid refresh token",
		ErrCodeInvalidCredentials:     "invalid username or password",
		ErrCodeUserIDRequired:         "userID is required",
		ErrCodeInternalError:          "internal server error",
		ErrCodeUsernameAlreadyExists:  "username already exists",
		ErrCodeFailedToRegister:       "failed to register user",
		ErrCodeFailedToLogin:          "failed to login",
		ErrCodeFailedToRefreshToken:   "failed to refresh token",
		ErrCodeFailedToGetUserInfo:    "failed to get user info",
		ErrCodeFailedToAddFriend:      "failed to add friend",
		ErrCodeFailedToGetFriendList:  "failed to get friend list",
		ErrCodeFailedToProcessRequest: "failed to process friend request",
		ErrCodeFailedToDeleteFriend:   "failed to delete friend",
		ErrCodeGroupNameEmpty:         "group name cannot be empty",
		ErrCodeFailedToCreateGroup:    "failed to create group",
		ErrCodeGroupNotFound:          "group not found",
		ErrCodeGroupAlreadyExists:     "group already exists",
		ErrCodePermissionDenied:       "permission denied",
		ErrCodeInvalidGroupRole:       "invalid group role",
		ErrCodeFailedToJoinGroup:      "failed to join group",
		ErrCodeFailedToLeaveGroup:     "failed to leave group",
		ErrCodeFailedToDisbandGroup:   "failed to disband group",
		ErrCodeFailedToTransferGroup:  "failed to transfer group",
		ErrCodeFailedToRemoveMember:   "failed to remove member",
		ErrCodeInviteCodeRequired:           "invite code required",
		ErrCodeInvalidInviteCode:            "invalid invite code",
		ErrCodeAlreadyInGroup:               "already in group",
		ErrCodeCannotLeaveAsOwner:           "cannot leave as owner",
		ErrCodeFailedToRequestJoinGroup:     "failed to request join group",
		ErrCodeJoinRequestNotFound:          "join request not found",
		ErrCodeFailedToApproveJoinRequest:    "failed to approve join request",
		ErrCodeCannotRequestWithinCooldown:  "cannot request within cooldown period",
		ErrCodeAlreadyRequested:             "already requested",
	}
)

func GetMessage(code int) string {
	if msg, ok := ErrMessages[code]; ok {
		return msg
	}
	return "unknown error"
}
