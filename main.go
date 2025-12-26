package main

import (
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/route"
)

func main() {

	// 加载配置
	config.Load()

	// 配置数据库
	database.InitDb()

	// 路由与中间件
	r := route.Route()

	// 服务启动
	if err := r.Run(":8080"); err != nil {
		return
	}

}
