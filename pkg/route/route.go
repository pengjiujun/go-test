package route

// 👇👇👇 添加这部分注释 👇👇👇
// @title           Go Gin Web 脚手架 API
// @version         1.0
// @description     这是一个基于 Gin 的后端 API 服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// 👆👆👆 添加这部分注释 👆👆👆
import (
	"github.com/gin-gonic/gin"
	"test/internal/controller"
	"test/internal/middleware"

	swaggerFiles "github.com/swaggo/files" // 👈 导入这两个包
	ginSwagger "github.com/swaggo/gin-swagger"
	// 👇 非常重要：必须导入刚才生成的 docs 包，路径要替换成你自己的项目模块名
	_ "test/docs"
)

func Route() *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.I18nMiddleware())
	gin.SetMode(gin.DebugMode)

	// 👇 添加这一行，注册 Swagger 路由接口
	// 访问 http://localhost:8080/swagger/index.html 即可看到文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")

	u := new(controller.UserController)
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
