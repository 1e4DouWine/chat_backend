package model

import (
	"time"

	"gorm.io/gorm"
)

type Friend struct {
	User1ID   string         `gorm:"type:uuid;not null;primaryKey"`
	User2ID   string         `gorm:"type:uuid;not null;primaryKey"`
	Status    string         `gorm:"type:text;not null;default:pending"` // accepted / pending
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
