package model

import "gorm.io/gorm"

type UrlRecord struct {
	gorm.Model
	OriginalUrl string
	ShortCode   string `gorm:"unique;not null"`
}

type URL struct {
	Url string `form:"url" json:"url" binding:"required"`
}
