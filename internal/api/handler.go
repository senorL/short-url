package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"short-url/internal/model"
	"short-url/internal/service"
	"short-url/pkg/base62"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type URLHandler struct {
	DB          *gorm.DB
	RDB         *redis.Client
	Leaf        *service.LeafNode
	Sg          singleflight.Group
	BloomFilter *bloom.BloomFilter
	LocalCache  *cache.Cache
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var urlJson model.URL
	if errJson := c.ShouldBindJSON(&urlJson); errJson != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errJson.Error()})
		return
	}
	id, err := h.Leaf.GetID()
	if err != nil {
		fmt.Println("发号失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发号器故障"})
		return
	}

	const magicPrime uint32 = 2654435761
	const xorMask uint32 = 95831234
	newID := (uint32(id) * magicPrime) ^ xorMask //确保shortcode看起来随机

	shortcode := base62.Base62Encode(uint64(newID))

	urlRecord := model.UrlRecord{OriginalUrl: urlJson.Url, ShortCode: shortcode}
	if err := h.DB.Create(&urlRecord).Error; err != nil {
		fmt.Println("数据库保存失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库保存失败"})
		return
	}
	h.RDB.Set(c.Request.Context(), "url:"+shortcode, urlJson.Url, 24*time.Hour)
	h.BloomFilter.AddString(shortcode)
	c.JSON(http.StatusOK, gin.H{
		"code":      http.StatusOK,
		"msg":       "success",
		"shortcode": shortcode,
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
	defer cancel()

	shortcode := c.Param("shortcode")

	if valueUrl, ok := h.LocalCache.Get(shortcode); ok {
		if valueUrl == "NULL" {
			c.String(http.StatusNotFound, "404页面迷路啦～")
			return
		}
		c.Redirect(http.StatusFound, valueUrl.(string))
		return
	}

	if !h.BloomFilter.TestString(shortcode) {
		c.String(http.StatusNotFound, "404页面迷路啦～ (被拦截)")
		return
	}

	url, err := h.RDB.Get(ctx, "url:"+shortcode).Result()
	if err == nil {
		if url == "NULL" {
			c.String(http.StatusNotFound, "404页面迷路啦～")
			return
		}

		c.Redirect(http.StatusFound, url)
	} else {
		v, err, _ := h.Sg.Do(shortcode, func() (interface{}, error) {
			var urlRecord model.UrlRecord
			result := h.DB.WithContext(ctx).Where("short_code = ?", shortcode).First(&urlRecord)

			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					h.RDB.Set(ctx, "url:"+shortcode, "NULL", 1*time.Minute)
				}
				return nil, result.Error
			}
			h.RDB.Set(ctx, "url:"+shortcode, urlRecord.OriginalUrl, 24*time.Hour)
			return urlRecord.OriginalUrl, nil
		})

		if err != nil {
			c.String(http.StatusNotFound, "404页面迷路啦～")
			return
		}

		c.Redirect(http.StatusFound, v.(string))
	}
}

func (h *URLHandler) GetLinks(c *gin.Context) {
	var urlRecords []model.UrlRecord
	h.DB.Find(&urlRecords)
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"data": urlRecords,
	})
}
