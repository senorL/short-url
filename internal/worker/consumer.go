package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"short-url/internal/model"
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func StartLogConsumer(db *gorm.DB) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "short-url-logger",
		Topic:    "url_access_logs",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  2 * time.Second,
	})
	defer r.Close()

	fmt.Println("Kafka启动...")
	var logBuffer []model.AccessLog

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("读取消息失败: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var logData model.AccessLog
		if err := json.Unmarshal(m.Value, &logData); err != nil {
			fmt.Printf("JSON解析失败: %v\n", err)
			continue
		}

		logBuffer = append(logBuffer, logData)

		if len(logBuffer) >= 50 {
			if err := db.Create(&logBuffer).Error; err != nil {
				fmt.Printf("批量插入SQL失败：%v\n", err)
			} else {
				fmt.Printf("成功导入%d条数据到SQL\n", len(logBuffer))
			}
			logBuffer = nil
		}

	}
}
