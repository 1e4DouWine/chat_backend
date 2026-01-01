package dto

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name string `json:"name"` // 群组名称，1-50字符
}

// GroupResponse 群组信息响应
type GroupResponse struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// GroupDetailResponse 群组详情响应
type GroupDetailResponse struct {
	GroupID     string            `json:"group_id"`
	Name        string            `json:"name"`
	OwnerID     string            `json:"owner_id"`
	OwnerName   string            `json:"owner_name"`
	MemberCount int               `json:"member_count"`
	CreatedAt   string            `json:"created_at"`
	Members     []GroupMemberInfo `json:"members"`
}

// GroupMemberInfo 群组成员信息
type GroupMemberInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

// GroupListResponse 群组列表项
type GroupListResponse struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	MemberCount int    `json:"member_count"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

// JoinGroupRequest 加入群组请求
type JoinGroupRequest struct {
	InviteCode string `json:"invite_code"` // 邀请码（可选）
}

// JoinGroupResponse 加入群组响应
type JoinGroupResponse struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
	Status  string `json:"status"` // joined
}

// TransferGroupRequest 转让群组请求
type TransferGroupRequest struct {
	NewOwnerID string `json:"new_owner_id"` // 新群主用户ID
}

// JoinGroupByCodeRequest 通过邀请码加入群组请求
type JoinGroupByCodeRequest struct {
	InviteCode string `json:"invite_code"` // 邀请码
}

// JoinGroupByCodeResponse 通过邀请码加入群组响应
type JoinGroupByCodeResponse struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
	Status  string `json:"status"` // joined
}

// SearchGroupResponse 搜索群组响应
type SearchGroupResponse struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}
