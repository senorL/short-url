package model

import "gorm.io/gorm"

type IDGenerator struct {
	gorm.Model
	MaxID uint64 `gorm:"not null"`
	Step  int    `gorm:"not null"`
}
