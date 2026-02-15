package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"test/internal/model"
	"test/pkg/config"
)

var DB *gorm.DB

func InitDb() {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Port,
		config.Conf.Database.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // 只打印错误，不打印警告
	})
	if err != nil {
		panic(err)
	}

	DB = db

	// 初始化数据表
	DB.AutoMigrate(
		&model.User{},
		&model.Banner{},
		&model.LmDtsGame{},
		&model.LmDtsRecord{},
	)

}
