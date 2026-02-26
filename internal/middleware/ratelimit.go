package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
if current > tonumber(ARGV[2]) then
    return 0
end
return 1
`)

func RateLimit(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		redisKey := "rate_limit:" + c.ClientIP()
		keys := []string{redisKey}
		values := []interface{}{int(window.Seconds()), maxRequests}
		result, err := rateLimitScript.Run(c.Request.Context(), rdb, keys, values...).Int()

		if err != nil {
			c.Next()
			return
		}
		if result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "过多次请求，请稍后再试"})
			return
		}
		c.Next()
	}
}
