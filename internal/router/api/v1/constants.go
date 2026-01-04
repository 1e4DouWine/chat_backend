package v1

const (
	QueryParamUsername     = "username"
	QueryParamStatus       = "status"
	QueryParamAction       = "action"
	QueryParamTargetUserID = "target_user_id"
	QueryParamLimit        = "limit"
	QueryParamCursor       = "cursor"
	QueryParamRole         = "role"
	QueryParamName         = "name"
	QueryParamRefreshToken = "refresh_token"

	ParamID      = "id"
	ParamGroupID = "group_id"
	ParamUserID  = "user_id"

	DefaultLimit        = 20
	DefaultHelloName    = "World"
	DefaultFriendStatus = "normal"

	ErrorMessageUsernameRequired          = "username is required"
	ErrorMessagePasswordRequired          = "password is required"
	ErrorMessageRefreshTokenRequired      = "refresh_token is required"
	ErrorMessageTargetUserIDRequired      = "target_user_id is required"
	ErrorMessageGroupIDRequired           = "group id is required"
	ErrorMessageNameRequired              = "name is required"
	ErrorMessageNewOwnerIDRequired        = "new_owner_id is required"
	ErrorMessageGroupIDAndUserIDRequired  = "group_id and user_id are required"
	ErrorMessageGroupIDAndUserIDRequired2 = "group id and user id are required"
	ErrorMessageInviteCodeRequired        = "invite_code is required"

	ErrorMessageCanNotSearchYourself  = "can not search yourself"
	ErrorMessageNotFriendRelationship = "You are not in a friend relationship"
	ErrorMessageNotGroupMember        = "You are not a member of this group"
	ErrorMessageYouAreNotInThisGroup  = "you are not in this group"

	ErrorMessageGroupNotFound               = "group not found"
	ErrorMessagePermissionDenied            = "permission denied"
	ErrorMessageAlreadyInGroup              = "already in group"
	ErrorMessageNotInGroup                  = "not in group"
	ErrorMessageCannotLeaveAsOwner          = "cannot leave as owner"
	ErrorMessageTargetUserNotInGroup        = "target user not in group"
	ErrorMessageCannotRemoveYourself        = "cannot remove yourself"
	ErrorMessageCannotRemoveOwner           = "cannot remove owner"
	ErrorMessageInvalidInviteCode           = "invalid invite code"
	ErrorMessagePendingRequestAlreadyExists = "pending request already exists"
	ErrorMessageCannotRequestWithinCooldown = "cannot request within cooldown period"
	ErrorMessageJoinRequestNotFound         = "join request not found"
	ErrorMessageInvalidAction               = "invalid action"

	ErrorMessageActionMustBeApproveOrReject = "action must be approve or reject"
)
