package main

import (
	"embed"
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/redis"
	"test/pkg/route"
	"test/pkg/translation"
)

//go:embed locales/*.toml
var rootLocales embed.FS

func main() {

	// 加载配置
	config.Load()

	// 配置数据库
	database.InitDb()

	redis.InitRedis()

	// 多语言翻译
	translation.InitComponents(rootLocales)

	// 路由与中间件
	r := route.Route()

	// 服务启动
	if err := r.Run(":8080"); err != nil {
		return
	}

}
