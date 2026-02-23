package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// var urlSwitch = make(map[string]string)
var count = 0

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
		c.String(http.StatusNotFound, "页面迷路啦～")
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
		count++
		shortcode := fmt.Sprintf("%d", count)
		urlRecord := UrlRecord{OriginalUrl: url, ShortCode: shortcode}
		db.Create(&urlRecord)
		c.HTML(http.StatusOK, "success.html", shortcode)

	})
	r.GET("/:shortcode", func(c *gin.Context) {
		// shortcode 转换成 url
		shortcode := c.Param("shortcode")
		var urlRecord UrlRecord
		db.Where("short_code = ?", shortcode).First(&urlRecord)
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
