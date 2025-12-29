package model

import "gorm.io/gorm"

type InvitationCode struct {
	gorm.Model
	UUID string `gorm:"type:uuid;not null"` // 群组的 ID 或用户的
	Name string `gorm:"type:text;not null"` // 群组的或用户的
	Code string `gorm:"type:text;not null;unique"`
}
