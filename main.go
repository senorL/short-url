package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type UrlRecord struct {
	gorm.Model
	OriginalUrl string
	ShortCode   string `gorm:"unique;not null"`
}

type URL struct {
	Url string `form:"url" json:"url" binding:"required"`
}

func main() {
	r := gin.Default()

	db, err := gorm.Open(sqlite.Open("short_url.db"), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败")
	}

	db.AutoMigrate(&UrlRecord{})

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404页面迷路啦～")
	})

	r.Use(StatCost())

	r.GET("/api/links", func(c *gin.Context) {
		var urlRecords []UrlRecord
		db.Find(&urlRecords)
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"msg":  "success",
			"data": urlRecords,
		})
	})
	r.POST("/shorten", func(c *gin.Context) {
		var urlJson URL
		err := c.ShouldBindJSON(&urlJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		shortcode := string(GenerateRandomCode(6))
		urlRecord := UrlRecord{OriginalUrl: urlJson.Url, ShortCode: shortcode}
		db.Create(&urlRecord)
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"msg":       "success",
			"shortcode": shortcode,
		})

	})
	r.GET("/:shortcode", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		// shortcode 转换成 url
		shortcode := c.Param("shortcode")
		var urlRecord UrlRecord
		result := db.WithContext(ctx).Where("short_code = ?", shortcode).First(&urlRecord)
		if result.Error != nil {
			c.String(http.StatusNotFound, "404页面迷路啦～")
			return
		}
		c.Redirect(http.StatusFound, urlRecord.OriginalUrl)
	})

	r.Run()
}

func StatCost() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cost := time.Since(start)

		fmt.Printf("请求 %s | 耗时 %v\n", c.Request.URL.Path, cost)
	}
}

func GenerateRandomCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}
