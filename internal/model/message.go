package model

import (
	"time"
)

type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

type Message struct {
	ID         string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FromUserID string      `gorm:"type:uuid;not null;index:idx_from;index:idx_private_chat"`
	TargetID   string      `gorm:"type:uuid;not null;index:idx_target;index:idx_private_chat"` // 用户ID或群ID，根据Type来判断
	Type       MessageType `gorm:"type:text;not null;index:idx_type"`
	Content    string      `gorm:"type:text;not null"`
	CreatedAt  time.Time   `gorm:"autoCreateTime;index:idx_created"`
}
