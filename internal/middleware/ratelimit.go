package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local clearBefore = now - window

redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

local count = redis.call('ZCARD', key)

if count >= limit then
    return 0 
else 
	redis.call('ZADD', key, now, now)
    redis.call('EXPIRE', key, window / 1000)
    return 1 
end
`)

func RateLimit(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		redisKey := "rate_limit:" + c.ClientIP()
		keys := []string{redisKey}
		values := []interface{}{time.Now().UnixMilli(), window.Milliseconds(), maxRequests}
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
