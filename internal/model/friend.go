package model

import (
	"time"

	"gorm.io/gorm"
)

type Friend struct {
	UserA            string         `gorm:"type:uuid;not null;primaryKey"`
	UserB            string         `gorm:"type:uuid;not null;primaryKey"`
	Status           string         `gorm:"type:text;not null;default:normal"` // normal（正常好友关系） / blacklist（其中一方将另一方加入了黑名单） / removed（解除好友关系）
	IsBlockedByUserA bool           `gorm:"type:boolean;not null;default:false"`
	IsBlockedByUserB bool           `gorm:"type:boolean;not null;default:false"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}
