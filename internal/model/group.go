package model

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"type:text;not null"`
	OwnerID     string         `gorm:"type:uuid;not null"`
	MemberCount int            `gorm:"type:int;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
