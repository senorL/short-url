package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func StatCost() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cost := time.Since(start)

		fmt.Printf("请求 %s | 耗时 %v\n", c.Request.URL.Path, cost)
	}
}
