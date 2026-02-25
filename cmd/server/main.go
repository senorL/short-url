package main

import (
	"net/http"
	"short-url/internal/api"
	"short-url/internal/middleware"
	"short-url/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
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

	db.AutoMigrate(&model.UrlRecord{})

	urlHandler := &api.URLHandler{
		DB:  db,
		RDB: rdb,
	}

	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404页面迷路啦～")
	})

	r.Use(middleware.StatCost())

	r.GET("/api/links", urlHandler.GetLinks)
	r.POST("/shorten", urlHandler.ShortenURL)
	r.GET("/:shortcode", urlHandler.Redirect)

	r.Run()
}
