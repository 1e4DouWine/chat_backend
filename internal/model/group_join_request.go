package model

import "gorm.io/gorm"

type GroupJoinRequest struct {
	gorm.Model
	SenderID      string `gorm:"type:uuid;not null;primaryKey"`
	TargetGroupID string `gorm:"type:uuid;not null;primaryKey"`
	Status        string `gorm:"type:text;not null;default:pending"` // only pending / rejected
	Message       string `gorm:"type:text"`
}
