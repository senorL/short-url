package main

import (
	"fmt"
	"html/template"
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

	r.SetFuncMap(template.FuncMap{
		"turl": TruncateURL,
	})

	r.LoadHTMLGlob("*.html")

	r.GET("/", func(c *gin.Context) {
		var urlRecords []UrlRecord
		db.Find(&urlRecords)
		c.HTML(http.StatusOK, "index.html", urlRecords)
	})
	r.POST("/shorten", func(c *gin.Context) {
		url := c.PostForm("url")
		shortcode := string(GenerateRandomCode(6))
		urlRecord := UrlRecord{OriginalUrl: url, ShortCode: shortcode}
		db.Create(&urlRecord)
		c.HTML(http.StatusOK, "success.html", shortcode)

	})
	r.GET("/:shortcode", func(c *gin.Context) {
		// shortcode 转换成 url
		shortcode := c.Param("shortcode")
		var urlRecord UrlRecord
		result := db.Where("short_code = ?", shortcode).First(&urlRecord)
		if result.Error != nil {
			c.String(http.StatusNotFound, "404页面迷路啦～")
		}
		c.Redirect(http.StatusFound, urlRecord.OriginalUrl)
	})

	r.Run()
}

func TruncateURL(url string) string {
	if len(url) > 30 {
		url = url[:30] + "..."
	}
	return url
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
