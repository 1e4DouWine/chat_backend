package dto

import "time"

type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

type MessageResponse struct {
	MessageID   string    `json:"message_id"`
	FromUserID  string    `json:"from_user_id"`
	TargetID    string    `json:"target_id"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	FromUser    *UserInfo `json:"from_user,omitempty"`
	TargetUser  *UserInfo `json:"target_user,omitempty"`
	TargetGroup *GroupInfo `json:"target_group,omitempty"`
}

type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type GroupInfo struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
}

type GetPrivateMessagesRequest struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
	Limit        int    `json:"limit"`
	Cursor       string `json:"cursor"`
}

type GetGroupMessagesRequest struct {
	GroupID string `json:"group_id" validate:"required"`
	Limit   int    `json:"limit"`
	Cursor  string `json:"cursor"`
}

type GetMessagesResponse struct {
	Messages    []MessageResponse `json:"messages"`
	NextCursor  string            `json:"next_cursor,omitempty"`
	HasMore     bool              `json:"has_more"`
}

type ConversationType string

const (
	ConversationTypePrivate ConversationType = "private"
	ConversationTypeGroup   ConversationType = "group"
)

type PrivateConversation struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Avatar       string    `json:"avatar"`
	LastContent  string    `json:"last_content"`
	LastTime     time.Time `json:"last_time"`
}

type GroupConversation struct {
	GroupID      string    `json:"group_id"`
	GroupName    string    `json:"group_name"`
	LastContent  string    `json:"last_content"`
	LastTime     time.Time `json:"last_time"`
	LastSenderID string    `json:"last_sender_id"`
	LastSenderName string  `json:"last_sender_name"`
}

type GetConversationListResponse struct {
	PrivateConversations []PrivateConversation `json:"private_conversations"`
	GroupConversations   []GroupConversation   `json:"group_conversations"`
}
