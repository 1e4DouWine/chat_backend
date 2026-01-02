package model

import "gorm.io/gorm"

type FriendRequest struct {
	gorm.Model
	SenderID   string `gorm:"type:uuid;not null;primaryKey"`
	ReceiverID string `gorm:"type:uuid;not null;primaryKey"`
	Status     string `gorm:"type:text;not null;default:pending"` // only pending / rejected
}
