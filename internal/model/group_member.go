package model

import (
	"time"

	"gorm.io/gorm"
)

type GroupMember struct {
	GroupID   string         `gorm:"type:uuid;not null;primaryKey"`
	UserID    string         `gorm:"type:uuid;not null;primaryKey"`
	Role      string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
