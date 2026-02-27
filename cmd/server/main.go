package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"short-url/internal/api"
	"short-url/internal/middleware"
	"short-url/internal/model"
	"short-url/internal/service"
	"short-url/internal/worker"
	"syscall"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"gorm.io/driver/mysql"
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

	err := godotenv.Load()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接 MySQL 失败: " + err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("获取数据库失败：" + err.Error())
	}

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db.AutoMigrate(&model.UrlRecord{})
	db.AutoMigrate(&model.IDGenerator{})
	db.AutoMigrate(&model.AccessLog{})

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

	lc := cache.New(5*time.Second, 10*time.Second)

	kafkaW := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  "url_access_logs",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	defer kafkaW.Close()

	urlHandler := &api.URLHandler{
		DB:          db,
		RDB:         rdb,
		Leaf:        leafNode,
		BloomFilter: bloomFilter,
		LocalCache:  lc,
		KafkaWriter: kafkaW,
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

	go worker.StartLogConsumer(db)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit //接收数据不赋值
	log.Println("服务关闭中...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
		log.Println("数据库连接已关闭")
	}

	log.Println("服务关闭")

}
