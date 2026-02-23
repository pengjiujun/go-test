package middleware

import "github.com/gin-gonic/gin"

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// 允许访问的域名，* 表示全部（生产环境建议指定具体域名）
			c.Header("Access-Control-Allow-Origin", origin)
			// 允许的 Header 头部字段
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
			// 允许的 HTTP 方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
			// 允许浏览器在跨域请求中携带凭证（如 Cookie）
			c.Header("Access-Control-Allow-Credentials", "true")
			// 设置缓存时间
			c.Header("Access-Control-Max-Age", "3600")
		}

		// 放行所有 OPTIONS 方法
		if method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
