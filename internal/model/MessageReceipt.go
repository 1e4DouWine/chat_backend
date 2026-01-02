package model

import "time"

// MessageReceipt 消息投递状态表
type MessageReceipt struct {
	MessageID   string    `gorm:"type:uuid;not null;primaryKey"`
	UserID      string    `gorm:"type:uuid;not null;primaryKey;index:idx_user_receipt"`
	IsDelivered bool      `gorm:"not null;default:false"` // 是否已送达（上线即标记为已送达）
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
