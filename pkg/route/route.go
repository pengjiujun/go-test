package route

import (
	"github.com/gin-gonic/gin"
	"test/internal/handle"
	"test/internal/middleware"
)

func Route() *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	gin.SetMode(gin.DebugMode)

	api := router.Group("/api")

	u := new(handle.User)
	user := api.Group("/user")
	{
		user.GET("/index", u.Index)
		user.POST("/create", u.Created)
		user.POST("/login", u.Login)
	}

	auth := api.Use(middleware.JWTAuth())
	{
		auth.GET("/user/show", u.Show)
	}

	return router
}
