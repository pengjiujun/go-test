package main

import (
	"context"
	"embed"
	"test/internal/process"
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

	// 2. 启动后台“自定义进程” (不阻塞，直接后台跑)
	// 这里的 DtsPushProcess 就是你刚才写的那个 select + ticker 逻辑
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go process.Handle(ctx)

	// 路由中间件
	r := route.Route()

	// 服务启动
	if err := r.Run(":8080"); err != nil {
		return
	}

}
