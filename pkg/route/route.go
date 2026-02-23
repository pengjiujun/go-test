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
	"test/pkg/config"
	app "test/pkg/jwt"

	swaggerFiles "github.com/swaggo/files" // 👈 导入这两个包
	ginSwagger "github.com/swaggo/gin-swagger"
	// 👇 非常重要：必须导入刚才生成的 docs 包，路径要替换成你自己的项目模块名
	_ "test/docs"
)

func Route() *gin.Engine {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.I18nMiddleware())
	router.Use(middleware.Cors())
	gin.SetMode(gin.DebugMode)

	// 1. 在这里初始化一次，单例使用
	jwtHandler := app.NewJWT(
		config.Conf.Jwt.Secret,
		config.Conf.Jwt.Issuer,
		config.Conf.Jwt.ExpireSeconds,
	)

	// 👇 添加这一行，注册 Swagger 路由接口
	// 访问 http://localhost:8080/swagger/index.html 即可看到文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 控制器初始化（建议放在一起，或者随用随开）
	//“模块前缀”与“鉴权逻辑”嵌套起来
	bannerCtrl := controller.NewBannerController()
	userCtrl := controller.NewUserController()
	dtsCtrl := controller.NewDtsController()

	v1 := router.Group("/api")
	{
		// 轮播图
		v1.GET("/banner", bannerCtrl.Index)

		// --- 用户模块 ---
		user := v1.Group("/user")
		{
			// 1. 无需授权的接口 (Public)
			user.GET("/index", userCtrl.Index)
			user.POST("/create", userCtrl.Created)
			user.POST("/login", userCtrl.Login)

			// 2. 需要授权的子组 (Private)
			// 嵌套一个子 Group，继承了 /user 前缀，并增加了 JWT 中间件
			userAuth := user.Group("/")
			userAuth.Use(middleware.JWTAuth(jwtHandler))
			{
				userAuth.GET("/show", userCtrl.Show) // 完整路径是 /api/user/show
			}
		}

		// --- 游戏模块 ---
		dts := v1.Group("/dts")
		{
			dts.GET("/ws", middleware.WsAuth(jwtHandler), dtsCtrl.Ws)

			dtsAuth := dts.Group("/")
			dtsAuth.Use(middleware.JWTAuth(jwtHandler))
			{
				dtsAuth.GET("/init", dtsCtrl.Init)  // 进入游戏
				dtsAuth.GET("/quit", dtsCtrl.Quit)  // 退出游戏
				dtsAuth.POST("/join", dtsCtrl.Join) // 加入游戏
			}
		}

	}
	return router
}
