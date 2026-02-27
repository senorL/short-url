package model

import (
	"time"

	"gorm.io/gorm"
)

type UrlRecord struct {
	gorm.Model
	OriginalUrl string
	ShortCode   string `gorm:"unique;not null"`
}

type URL struct {
	Url string `form:"url" json:"url" binding:"required"`
}

type AccessLog struct {
	ShortCode string    `json:"short_code"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}
