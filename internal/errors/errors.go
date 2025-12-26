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
	}
)

func GetMessage(code int) string {
	if msg, ok := ErrMessages[code]; ok {
		return msg
	}
	return "unknown error"
}
