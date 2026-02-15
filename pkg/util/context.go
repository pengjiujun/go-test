package util

import "github.com/gin-gonic/gin"

func GetUserID(c *gin.Context) int64 {
	value, exists := c.Get("userID")
	if !exists {
		return 0
	}
	// 增加防御性断言
	id, ok := value.(int64)
	if !ok {
		return 0
	}
	return id
}
