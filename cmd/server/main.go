package main

import (
	"net/http"
	"short-url/internal/api"
	"short-url/internal/middleware"
	"short-url/internal/model"
	"short-url/internal/service"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	ginprometheus "github.com/zsais/go-gin-prometheus"
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
	db.AutoMigrate(&model.IDGenerator{})

	db.FirstOrCreate(&model.IDGenerator{
		MaxID: 15000000,
		Step:  1000,
	}, model.IDGenerator{Model: gorm.Model{ID: 1}})

	leafNode, err := service.NewLeafNode(db)
	if err != nil {
		panic("发号器启动失败" + err.Error())
	}

	bloomFilter := bloom.NewWithEstimates(10000000, 0.0001)
	var allShortCodes []string
	db.Model(&model.UrlRecord{}).Pluck("short_code", &allShortCodes)
	for _, code := range allShortCodes {
		bloomFilter.AddString(code)
	}

	urlHandler := &api.URLHandler{
		DB:          db,
		RDB:         rdb,
		Leaf:        leafNode,
		BloomFilter: bloomFilter,
	}

	r := gin.Default()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404页面迷路啦～")
	})

	r.Use(middleware.StatCost())

	r.GET("/api/links", urlHandler.GetLinks)
	r.POST("/shorten", middleware.RateLimit(rdb, 10, 5*time.Second), urlHandler.ShortenURL)
	r.GET("/:shortcode", urlHandler.Redirect)

	r.Run()
}
