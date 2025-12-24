package model

import (
	"time"

	"gorm.io/gorm"
)

type GroupMember struct {
	GroupID   string         `gorm:"type:uuid;not null;primaryKey"`
	UserID    string         `gorm:"type:uuid;not null;primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
