package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var urlSwitch = make(map[string]string)
var count = 0

func main() {
	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "页面迷路啦～")
	})

	r.Use(StatCost())

	r.SetFuncMap(template.FuncMap{
		"turl": TruncateURL,
	})

	r.LoadHTMLGlob("*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", urlSwitch)
	})
	r.POST("/shorten", func(c *gin.Context) {
		url := c.PostForm("url")
		count++
		shortcode := fmt.Sprintf("%d", count)
		urlSwitch[shortcode] = url
		c.HTML(http.StatusOK, "success.html", shortcode)

	})
	r.GET("/:shortcode", func(c *gin.Context) {
		// shortcode 转换成 url
		shortcode := c.Param("shortcode")
		url, ok := urlSwitch[shortcode]
		if ok {
			c.Redirect(http.StatusFound, url)
		} else {
			c.String(http.StatusNotFound, "链接不存在")
		}

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
