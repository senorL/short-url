package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const base62Charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	defer rdb.Close()

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
		if errJson := c.ShouldBindJSON(&urlJson); errJson != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errJson.Error()})
			return
		}
		id, err := rdb.Incr(c.Request.Context(), "short_url_id").Result()
		if err != nil {
			fmt.Println("Redis 发号失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis发号器故障"})
			return
		}

		shortcode := Base62Encode(uint64(id))
		urlRecord := UrlRecord{OriginalUrl: urlJson.Url, ShortCode: shortcode}
		if err := db.Create(&urlRecord).Error; err != nil {
			fmt.Println("数据库保存失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库保存失败"})
			return
		}
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

func Base62Encode(id uint64) string {
	var code []byte
	if id == 0 {
		code = append(code, '0')
		return string(code)
	}
	for id > 0 {
		code = append(code, base62Charset[id%62])
		id /= 62
	}
	slices.Reverse(code)
	return string(code)
}
