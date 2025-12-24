package dto

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	Username string `json:"username"` // 目标用户名
	Message  string `json:"message"`  // 好友申请消息
}

// AddFriendResponse 好友申请响应
type AddFriendResponse struct {
	FriendID string `json:"friend_id"`
	Status   string `json:"status"` // pending | accepted
}
