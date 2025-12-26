package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"test/pkg/config"
	app "test/pkg/jwt" // 导入刚才的包
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		tokenHeader := c.Request.Header.Get("Authorization")
		if tokenHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请求未携带Token，无权访问"})
			c.Abort()
			return
		}

		// 2. 格式校验 "Bearer <token>"
		parts := strings.SplitN(tokenHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token格式错误"})
			c.Abort()
			return
		}

		// 3. 初始化 JWT 实例
		jwtHandler := app.NewJWT(config.Conf.Jwt.Secret, config.Conf.Jwt.Issuer, config.Conf.Jwt.ExpireSeconds)

		// 4. 解析 Token
		claims, err := jwtHandler.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
			c.Abort()
			return
		}

		// 5. 关键步骤：将用户信息存入上下文 (Context)
		// 之后所有的 Controller 都能直接从 c 里拿到 UserID
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
